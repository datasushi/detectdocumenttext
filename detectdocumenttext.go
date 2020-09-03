package main

import (
	"bufio"
	"image"
	"math"
	"os"
	
	"image/color"
	
	"image/png"
	"io/ioutil"
	
    "go.uber.org/zap"

    "golang.org/x/image/font"
    "golang.org/x/image/font/basicfont"
    "golang.org/x/image/math/fixed"
    
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/textract"
)

var logger *zap.Logger
var imgout image.RGBA

func main() {
//loggger, _ := zap.NewDevelopment()
//defer loggger.Sync()
//logger = loggger.Sugar()
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	logger.Info("It's alive",
		zap.String("Vitality test:", "passed"),
	
	)
	
	sess := session.New(&aws.Config{
  		Region: aws.String("us-east-2"),
	})
	svc := textract.New(sess)
	
	fileName := "all.png"
	f, _ := os.Open(fileName)
	reader := bufio.NewReader(f)
	content, _ := ioutil.ReadAll(reader)
	detectDocumentTextResult, _ := svc.DetectDocumentText(
		&textract.DetectDocumentTextInput{
			Document: &textract.Document {
				Bytes: content,
	}})

	f, _ = os.Open(fileName)
	img, _, _ := image.DecodeConfig(f)
	width := img.Width
	height := img.Height
	logger.Info("Detected Image Config",
		zap.Int("Width: ", width),
		zap.Int("Height: ", height),
	)
	upLeft := image.Point{0, 0}
	lowRight := image.Point{img.Width, img.Height}
	imgout := image.NewRGBA(image.Rectangle{upLeft, lowRight})
	whiten(imgout, width, height)

	wordCount := 1

	for _, detectedBlock := range detectDocumentTextResult.Blocks {	
		logger.Info("Block detected",
			zap.String("Type: ", *detectedBlock.BlockType),
		)

		if *detectedBlock.BlockType == "WORD" {
			logger.Info("Block detected",
				zap.String("Type: ", *detectedBlock.BlockType),
				zap.Int("Word Count: ", wordCount),
			)
			wordCount++;
			y1 := int(math.Round(*detectedBlock.Geometry.BoundingBox.Top*float64(height)))
			x1 := int(math.Round(*detectedBlock.Geometry.BoundingBox.Left*float64(width)))
			y2 := y1 + int(math.Round(*detectedBlock.Geometry.BoundingBox.Height*float64(height)))
			x2 := x1 + int(math.Round(*detectedBlock.Geometry.BoundingBox.Width*float64(width)))
			Rect(imgout, x1, y1, x2, y2)
			addLabel(imgout, x1+10, y1+20, *detectedBlock.Text)
		}
	}
	fout, _ := os.Create("out.png")
	png.Encode(fout, imgout)

}

func evaluateError(err error) {
	if err != nil {
		logger.Error("Critical Error",
			zap.Error(err),
		)
	}
}

func addLabel(img *image.RGBA, x, y int, label string) {
    point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}

    d := &font.Drawer{
        Dst:  img,
        Src:  image.NewUniform(color.Black),
        Face: basicfont.Face7x13,
        Dot:  point,
    }
    d.DrawString(label)
}

func whiten(img *image.RGBA, w int, h int) {
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			img.Set(x, y, color.White)
		}
	}
}

func HLine(img *image.RGBA, x1, y, x2 int) {
    for ; x1 <= x2; x1++ {
        img.Set(x1, y, color.Black)
    }
}

// VLine draws a veritcal line
func VLine(img *image.RGBA, x, y1, y2 int) {
    for ; y1 <= y2; y1++ {
        img.Set(x, y1, color.Black)
    }
}

// Rect draws a rectangle utilizing HLine() and VLine()
func Rect(img *image.RGBA, x1, y1, x2, y2 int) {
    HLine(img, x1, y1, x2)
    HLine(img, x1, y2, x2)
    VLine(img, x1, y1, y2)
    VLine(img, x2, y1, y2)
}
