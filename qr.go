package main

import (
	"bytes"
	"log"
	"net/url"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

type bufferWriteCloser struct {
	*bytes.Buffer
}

func (bwc *bufferWriteCloser) Close() error {
	return nil // No actual resources to release
}

func claim_score_qr(codeID string) {
	// Build base URL
	baseURL := &url.URL{
		Scheme: "http",
		Host:   remote_ip,
		Path:   "/claim",
	}

	// Add query parameters
	query := url.Values{}
	query.Set("code", codeID)
	baseURL.RawQuery = query.Encode()

	// Create qr coed
	qrc, err := qrcode.New(baseURL.String())
	if err != nil {
		log.Println("Failed to create qr code")
		log.Println(err)
		return
	}

	// Save qrcode
	var buf bytes.Buffer
	bwc := &bufferWriteCloser{&buf}
	w := standard.NewWithWriter(bwc)
	err = qrc.Save(w)
	if err != nil {
		log.Println("Failed to save qr code")
		log.Println(err)
		return
	}
	qr_code = buf.Bytes()
}

func blank_qr() {
	// Create qr coed
	qrc, err := qrcode.New("http://" + remote_ip)
	if err != nil {
		log.Println("Failed to create qr code")
		log.Println(err)
		return
	}

	// Save qrcode
	var buf bytes.Buffer
	bwc := &bufferWriteCloser{&buf}
	w := standard.NewWithWriter(bwc)
	err = qrc.Save(w)
	if err != nil {
		log.Println("Failed to save qr code")
		log.Println(err)
		return
	}
	qr_code = buf.Bytes()
}
