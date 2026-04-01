package service

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

type smtpConfig struct {
	host          string
	port          int
	username      string
	password      string
	from          string
	useSSL        bool
	skipTLSVerify bool
	timeout       time.Duration
	forceIPv4     bool
}

func loadSMTPConfig() (smtpConfig, error) {
	port, err := strconv.Atoi(defaultString(os.Getenv("SMTP_PORT"), "587"))
	if err != nil {
		return smtpConfig{}, fmt.Errorf("invalid SMTP_PORT: %w", err)
	}

	timeoutSeconds, err := strconv.Atoi(defaultString(os.Getenv("SMTP_TIMEOUT_SEC"), "5"))
	if err != nil {
		return smtpConfig{}, fmt.Errorf("invalid SMTP_TIMEOUT_SEC: %w", err)
	}

	cfg := smtpConfig{
		host:          os.Getenv("SMTP_HOST"),
		port:          port,
		username:      os.Getenv("SMTP_USERNAME"),
		password:      os.Getenv("SMTP_PASSWORD"),
		from:          defaultString(os.Getenv("SMTP_FROM"), os.Getenv("SMTP_USERNAME")),
		useSSL:        port == 465 || strings.EqualFold(os.Getenv("SMTP_SSL"), "true"),
		skipTLSVerify: strings.EqualFold(os.Getenv("SMTP_SKIP_TLS_VERIFY"), "true"),
		timeout:       time.Duration(timeoutSeconds) * time.Second,
		forceIPv4:     !strings.EqualFold(os.Getenv("SMTP_FORCE_IPV4"), "false"),
	}

	if cfg.host == "" || cfg.username == "" || cfg.password == "" || cfg.from == "" {
		return smtpConfig{}, errors.New("SMTP credentials are not configured")
	}

	return cfg, nil
}

func smtpAuthCheck(ctx context.Context) error {
	cfg, err := loadSMTPConfig()
	if err != nil {
		return err
	}

	client, closeFn, err := dialSMTPClient(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeFn()

	return client.Quit()
}

func sendSMTPTextEmail(ctx context.Context, to, subject, body string) error {
	cfg, err := loadSMTPConfig()
	if err != nil {
		return err
	}

	client, closeFn, err := dialSMTPClient(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeFn()

	if err := client.Mail(cfg.from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}

	message := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s\r\n", cfg.from, to, subject, body)
	if _, err := writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return client.Quit()
}

func dialSMTPClient(ctx context.Context, cfg smtpConfig) (*smtp.Client, func(), error) {
	addresses, network, err := resolveSMTPAddresses(ctx, cfg.host, cfg.port, cfg.forceIPv4)
	if err != nil {
		return nil, func() {}, err
	}
	tlsConfig := &tls.Config{
		ServerName:         cfg.host,
		InsecureSkipVerify: cfg.skipTLSVerify,
	}

	if cfg.useSSL {
		var lastErr error
		for _, addr := range addresses {
			timeout, err := remainingTimeout(ctx, cfg.timeout)
			if err != nil {
				return nil, func() {}, err
			}
			dialer := &net.Dialer{Timeout: timeout}

			conn, err := tls.DialWithDialer(dialer, network, addr, tlsConfig)
			if err != nil {
				lastErr = err
				continue
			}
			_ = conn.SetDeadline(time.Now().Add(timeout))

			client, err := smtp.NewClient(conn, cfg.host)
			if err != nil {
				_ = conn.Close()
				lastErr = err
				continue
			}
			if err := client.Auth(smtp.PlainAuth("", cfg.username, cfg.password, cfg.host)); err != nil {
				_ = client.Close()
				_ = conn.Close()
				lastErr = err
				continue
			}

			return client, func() {
				_ = client.Close()
				_ = conn.Close()
			}, nil
		}

		if lastErr == nil {
			lastErr = errors.New("SMTP connection failed")
		}
		return nil, func() {}, lastErr
	}

	var lastErr error
	for _, addr := range addresses {
		timeout, err := remainingTimeout(ctx, cfg.timeout)
		if err != nil {
			return nil, func() {}, err
		}
		dialer := &net.Dialer{Timeout: timeout}

		conn, err := dialer.DialContext(ctx, network, addr)
		if err != nil {
			lastErr = err
			continue
		}
		_ = conn.SetDeadline(time.Now().Add(timeout))

		client, err := smtp.NewClient(conn, cfg.host)
		if err != nil {
			_ = conn.Close()
			lastErr = err
			continue
		}

		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsConfig); err != nil {
				_ = client.Close()
				_ = conn.Close()
				lastErr = err
				continue
			}
		}

		if err := client.Auth(smtp.PlainAuth("", cfg.username, cfg.password, cfg.host)); err != nil {
			_ = client.Close()
			_ = conn.Close()
			lastErr = err
			continue
		}

		return client, func() {
			_ = client.Close()
			_ = conn.Close()
		}, nil
	}

	if lastErr == nil {
		lastErr = errors.New("SMTP connection failed")
	}
	return nil, func() {}, lastErr
}

func resolveSMTPAddresses(ctx context.Context, host string, port int, ipv4Only bool) ([]string, string, error) {
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, "", err
	}

	addresses := make([]string, 0, len(ips))
	network := "tcp"
	if ipv4Only {
		network = "tcp4"
	}

	for _, ip := range ips {
		if ipv4Only && ip.IP.To4() == nil {
			continue
		}
		addresses = append(addresses, net.JoinHostPort(ip.IP.String(), strconv.Itoa(port)))
	}

	if len(addresses) == 0 {
		return nil, "", fmt.Errorf("no SMTP addresses resolved for %s", host)
	}

	return addresses, network, nil
}

func remainingTimeout(ctx context.Context, fallback time.Duration) (time.Duration, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	timeout := fallback
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return 0, context.DeadlineExceeded
		}
		if remaining < timeout {
			timeout = remaining
		}
	}

	return timeout, nil
}
