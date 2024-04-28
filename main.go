package main

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"image"
	"io"
	"os"

	"image/color"
	"image/jpeg"
	"image/png"

	"github.com/valyala/fasthttp"
)

func faviconHandler(ctx *fasthttp.RequestCtx) {

	fileData, _ := os.ReadFile("./public/favicon.png")
	ctx.SetContentType("image/png")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Write(fileData)
}

func indexHandler(ctx *fasthttp.RequestCtx) {

	fileData, _ := os.ReadFile("./public/index.html")
	ctx.SetContentType("text/html")
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Write(fileData)
}

func uploadHandler(ctx *fasthttp.RequestCtx) {

	h, err := ctx.FormFile("file")
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		fmt.Fprintf(ctx, "error retrieving file: %v\n", err)
		return
	}

	uploadFile, err := h.Open()
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		fmt.Fprintf(ctx, "error opening file: %v\n", err)
		return
	}
	defer uploadFile.Close()
	fileReader := io.ReadSeeker(uploadFile)

	_, format, err := image.DecodeConfig(fileReader)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		fmt.Fprintf(ctx, "error unknown file type: %v\n", err)
		return
	}
	fileReader.Seek(0, 0)

	if format == "png" {
		pngImg, err := png.Decode(fileReader)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			fmt.Fprintf(ctx, "error server: %v\n", err)
			return
		}

		img := image.NewNRGBA(pngImg.Bounds())
		for y := 0; y < pngImg.Bounds().Dy(); y++ {
			for x := 0; x < pngImg.Bounds().Dx(); x++ {
				r, g, b, a := pngImg.At(x, y).RGBA()
				img.SetRGBA64(x, y, color.RGBA64{
					R: uint16(r),
					G: uint16(g),
					B: uint16(b),
					A: uint16(a),
				})
			}
		}

		buff := bytes.NewBuffer([]byte(""))

		if err = jpeg.Encode(buff, img, nil); err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			fmt.Fprintf(ctx, "error server: %v\n", err)
			return
		}

		fileReader = io.ReadSeeker(bytes.NewReader(buff.Bytes()))
	} else {
		if format != "jpeg" {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			fmt.Fprintf(ctx, "error unknown file type: %v\n", err)
			return
		}
	}

	hash := md5.New()
	if _, err := io.Copy(hash, fileReader); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		fmt.Fprintf(ctx, "error server: %v\n", err)
		return
	}
	fileReader.Seek(0, 0)

	filehash := fmt.Sprintf("%x", hash.Sum(nil))
	f, err := os.Create(fmt.Sprintf("images/%s.jpeg", filehash))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		fmt.Fprintf(ctx, "error server: %v\n", err)
		return
	}

	if _, err := io.Copy(f, fileReader); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		fmt.Fprintf(ctx, "error server: %v\n", err)
		return
	}

	ctx.SetContentType("text/html")
	fmt.Fprintf(ctx, `<a href="/%s">%s</a>`, filehash, filehash)
}

func imageHandler(ctx *fasthttp.RequestCtx) {

	file, err := os.Open(fmt.Sprintf("images%s.jpeg", ctx.Path()))
	if err != nil {
		ctx.NotFound()
		// ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		// fmt.Fprintf(ctx, "error server: %v\n", err)
		return
	}

	ctx.SetContentType("image/jpeg")
	if _, err := io.Copy(ctx, file); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		fmt.Fprintf(ctx, "error server: %v\n", err)
		return
	}
}

func router(ctx *fasthttp.RequestCtx) {

	method := string(ctx.Method())
	path := string(ctx.Path())

	switch method + ":" + path {

	// static files
	case "GET:/favicon.ico":
		faviconHandler(ctx)
	case "GET:/":
		indexHandler(ctx)

	// image upload
	case "POST:/upload":
		uploadHandler(ctx)

	default:
		imageHandler(ctx)
	}
}

func main() {

	server := &fasthttp.Server{
		Handler:            router,
		MaxRequestBodySize: 1024 * 1024 * 10, // 10mb
	}

	if err := server.ListenAndServe(":8080"); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
