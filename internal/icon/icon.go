package icon

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
)

// Data contains the icon bytes
var Data []byte

func init() {
	Data = createIcon()
}

func createIcon() []byte {
	// Create 16x16 image
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))

	// Orange color for cloud
	orange := color.RGBA{255, 140, 0, 255}

	// Draw simple cloud shape
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			// Cloud shape
			if (y >= 4 && y <= 11 && x >= 2 && x <= 13) ||
				(y >= 3 && y <= 12 && x >= 4 && x <= 11) ||
				(y >= 5 && y <= 10 && x >= 1 && x <= 14) {
				img.Set(x, y, orange)
			}
		}
	}

	// Convert to ICO format
	return toICO(img)
}

func toICO(img *image.RGBA) []byte {
	buf := new(bytes.Buffer)

	// ICO header
	binary.Write(buf, binary.LittleEndian, uint16(0))     // Reserved
	binary.Write(buf, binary.LittleEndian, uint16(1))     // Type: ICO
	binary.Write(buf, binary.LittleEndian, uint16(1))     // Count

	// ICO directory entry
	buf.WriteByte(16)                                      // Width
	buf.WriteByte(16)                                      // Height
	buf.WriteByte(0)                                       // Colors
	buf.WriteByte(0)                                       // Reserved
	binary.Write(buf, binary.LittleEndian, uint16(1))     // Planes
	binary.Write(buf, binary.LittleEndian, uint16(32))    // BPP

	// Calculate image data size
	imageDataSize := uint32(40 + 16*16*4 + 16*4) // header + pixels + mask
	binary.Write(buf, binary.LittleEndian, imageDataSize)
	binary.Write(buf, binary.LittleEndian, uint32(22))    // Offset

	// BITMAPINFOHEADER
	binary.Write(buf, binary.LittleEndian, uint32(40))    // Header size
	binary.Write(buf, binary.LittleEndian, int32(16))     // Width
	binary.Write(buf, binary.LittleEndian, int32(32))     // Height (doubled for ICO)
	binary.Write(buf, binary.LittleEndian, uint16(1))     // Planes
	binary.Write(buf, binary.LittleEndian, uint16(32))    // BPP
	binary.Write(buf, binary.LittleEndian, uint32(0))     // Compression
	binary.Write(buf, binary.LittleEndian, uint32(0))     // Image size
	binary.Write(buf, binary.LittleEndian, int32(0))      // X ppm
	binary.Write(buf, binary.LittleEndian, int32(0))      // Y ppm
	binary.Write(buf, binary.LittleEndian, uint32(0))     // Colors used
	binary.Write(buf, binary.LittleEndian, uint32(0))     // Important colors

	// Pixel data (bottom-up, BGRA)
	for y := 15; y >= 0; y-- {
		for x := 0; x < 16; x++ {
			c := img.RGBAAt(x, y)
			buf.WriteByte(c.B)
			buf.WriteByte(c.G)
			buf.WriteByte(c.R)
			buf.WriteByte(c.A)
		}
	}

	// AND mask (all zeros = all visible)
	for i := 0; i < 16*4; i++ {
		buf.WriteByte(0)
	}

	return buf.Bytes()
}
