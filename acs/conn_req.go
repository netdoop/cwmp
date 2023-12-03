package acs

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

func (s *AcsServer) SendConnectRequest(device Device) error {
	url := device.MustGetParameterValue("Device.ManagementServer.ConnectionRequestURL")
	username := device.MustGetParameterValue("Device.ManagementServer.ConnectionRequestUsername")
	password := device.MustGetParameterValue("Device.ManagementServer.ConnectionRequestPassword")
	udpConnectionRequestAddress := device.MustGetParameterValue("Device.ManagementServer.UDPConnectionRequestAddress")
	natDetected := device.MustGetParameterValue("Device.ManagementServer.NATDetected")

	if url != "" {
		ok, err := sendHttpConnetionRequest(url, username, password)
		if err != nil {
			return errors.Wrap(err, "send http connect request")
		}
		if ok {
			return nil
		}
	}

	if natDetected == "1" && udpConnectionRequestAddress != "" {
		for i := 0; i < 3; i++ {
			sendUDPConnectionRequest(udpConnectionRequestAddress, username, password)
			time.Sleep(time.Second)
		}
	}
	return nil
}

func sendHttpConnetionRequest(url string, username string, password string) (bool, error) {
	if url == "" {
		return false, nil
	}
	if isLANAddress(url) {
		return false, nil
	}
	client := &http.Client{
		Timeout: time.Second * 10, // Response timeout of 10 seconds
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   2 * time.Second, // Connection timeout of 2 seconds
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, errors.Wrap(err, "create request")
	}
	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		return false, errors.Wrap(err, "send request")
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return true, nil
	}
	return false, errors.Errorf("connection request failed. Status code: %v", resp.StatusCode)
}

func sendUDPConnectionRequest(addr string, username string, password string) (bool, error) {
	if addr == "" {
		return false, nil
	}
	now := time.Now()
	timestamp := now.Unix()
	messageID := fmt.Sprintf("%v", now.UnixNano())
	cnonce := generateRandomString(16)

	conn, err := net.Dial("udp", addr)
	if err != nil {
		return false, errors.Wrap(err, "create udp connection")
	}
	defer conn.Close()

	message := fmt.Sprintf("GET http://%s?ts=%d&id=%s&un=%s&cn=%s&sig=%s HTTP/1.1\r\n"+
		"Host: %s\r\n"+
		"Connection: close\r\n"+
		"\r\n",
		addr,
		timestamp, messageID, username, cnonce,
		calculateSignature(timestamp, messageID, username, cnonce, password),
		addr)

	// Send the UDP Connection Request message
	_, err = conn.Write([]byte(message))
	if err != nil {
		return false, errors.Wrap(err, "send udp connection request")
	}
	return true, nil
}

func calculateSignature(timestamp int64, messageID, username, cnonce, password string) string {
	text := fmt.Sprintf("%d%s%s%s", timestamp, messageID, username, cnonce)
	key := []byte(password)
	h := hmac.New(sha1.New, key)
	h.Write([]byte(text))
	signature := fmt.Sprintf("%x", h.Sum(nil))
	return signature
}

func isLANAddress(host string) bool {
	lanBlocks := []string{
		"192.168.0.0/16",
		"172.16.0.0/12",
		"10.0.0.0/8",
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return false
	}
	for _, ip := range ips {
		for _, block := range lanBlocks {
			_, subnet, _ := net.ParseCIDR(block)
			if subnet.Contains(ip) {
				return true
			}
		}
	}
	return false
}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func generateRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	randomString := make([]byte, length)
	for i := 0; i < length; i++ {
		randomString[i] = charset[rand.Intn(len(charset))]
	}
	return string(randomString)
}
