package notifier_test

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"gopuppy/internal/notifier"
)

func TestSMTPSenderCompletesDataBeforeQuit(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	done := make(chan error, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(500 * time.Millisecond))
		r := bufio.NewReader(conn)
		w := bufio.NewWriter(conn)
		write := func(s string) error {
			if _, err := fmt.Fprint(w, s+"\r\n"); err != nil {
				return err
			}
			return w.Flush()
		}
		readPrefix := func(prefix string) error {
			line, err := r.ReadString('\n')
			if err != nil {
				return err
			}
			if !strings.HasPrefix(line, prefix) {
				return fmt.Errorf("expected %s, got %q", prefix, line)
			}
			return nil
		}

		if err := write("220 localhost ESMTP"); err != nil {
			done <- err
			return
		}
		for _, step := range []struct {
			command, response string
		}{{"EHLO ", "250 localhost"}, {"MAIL FROM:", "250 ok"}, {"RCPT TO:", "250 ok"}, {"DATA", "354 end data"}} {
			if err := readPrefix(step.command); err != nil {
				done <- err
				return
			}
			if err := write(step.response); err != nil {
				done <- err
				return
			}
		}
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				done <- err
				return
			}
			if line == ".\r\n" {
				break
			}
		}
		if err := write("250 queued"); err != nil {
			done <- err
			return
		}
		if err := readPrefix("QUIT"); err != nil {
			done <- err
			return
		}
		done <- write("221 bye")
	}()

	host, portText, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	sender := &notifier.SMTPSender{Host: host, Port: port, From: "alerts@gopuppy.test"}
	err = sender.Send(context.Background(), notifier.Message{
		ToEmail: "owner@gopuppy.test",
		Title:   "体检提醒",
		Body:    "该带小狗去体检了",
	})
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if err := <-done; err != nil {
		t.Fatalf("SMTP session error = %v", err)
	}
}
