package qr

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/cyberark/conjur-cli-go/pkg/cmd/style"
	"image/png"
	"strings"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	qr "github.com/skip2/go-qrcode"
)

func DisplayQRCode(img string) error {
	base64Data := strings.TrimPrefix(img, "data:image/png;base64,")
	imgBytes, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return fmt.Errorf("failed to decode base64 image: %w", err)
	}
	iImg, err := png.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return fmt.Errorf("failed to decode PNG image: %w", err)
	}

	bmp, err := gozxing.NewBinaryBitmapFromImage(iImg)
	if err != nil {
		return fmt.Errorf("failed to create binary bitmap: %w", err)
	}

	qrReader := qrcode.NewQRCodeReader()
	codes, err := qrReader.Decode(bmp, nil)
	if err != nil {
		return fmt.Errorf("failed to decode QR Code: %w", err)
	}

	qrCode, err := qr.New(codes.String(), qr.Highest)
	if err != nil {
		return fmt.Errorf("failed to create QR Code: %w", err)
	}

	// The CyberArk authenticator app QR scanner requires color is inverted on light backgrounds
	fmt.Println(qrCode.ToSmallString(!style.HasDarkBackground()))
	return nil
}
