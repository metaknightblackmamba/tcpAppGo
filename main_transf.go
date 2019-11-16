package main

import (
    "image"
    "image/color"
    "image/png"
    "image/jpeg"
    "image/draw"
    "log"
    "os"
)


func TransformToGrey(x int, h int, imgsrc image.Image, imgdst *image.Gray){

  for y := 0; y < h; y++ {

      imageColor := imgsrc.At(x, y)
      rr, gg, bb, _ := imageColor.RGBA()
      Y := uint16((rr + gg + bb) / 3)
      grayColor := color.Gray{uint8(Y >> 8)}
      imgdst.Set(x, y, grayColor)

    }

}







func main() {
    filename := os.Args[1]
    infile, err := os.Open(filename)

    if err != nil {
        log.Printf("failed opening %s: %s", filename, err)
        panic(err.Error())
    }
    defer infile.Close()

    imgSrc, _, err := image.Decode(infile)
    if err != nil {
        panic(err.Error())
    }

    // Create a new grayscale image
    bounds := imgSrc.Bounds()
    w, h := bounds.Max.X, bounds.Max.Y
    grayScale := image.NewGray(image.Rectangle{image.Point{0, 0}, image.Point{w, h}})

    for x := 0; x < w; x++ {

      go TransformToGrey(x, h, imgSrc, grayScale)


    }

    // Encode the grayscale image to the new file
    newFileName := "grayscale.png"
    newfile, err := os.Create(newFileName)
    if err != nil {
        log.Printf("failed creating %s: %s", newfile, err)
        panic(err.Error())
    }
    defer newfile.Close()
    png.Encode(newfile,grayScale)


    if len(os.Args) > 1 && os.Args[2]=="Licorne" {

    image1, err := os.Open(os.Args[1])
    if err != nil {
        log.Fatalf("failed to open: %s", err)
    }
     
    first, err := png.Decode(image1)
    if err != nil {
        log.Fatalf("failed to decode: %s", err)
    }
    defer image1.Close()
 
    image2, err1 := os.Open("licorne1.png")
    if err1 != nil {
        log.Fatalf("failed to open: %s", err)
    }
    second,err := png.Decode(image2)
    if err != nil {
        log.Fatalf("failed to decode: %s", err)
    }
    defer image2.Close()
    
    unicorne(first, second)
    } 
    
}


func unicorne(picture1 image.Image, picture2 image.Image) {
        imgRef := picture1.Bounds()
        imgLic := picture2.Bounds()
        offset := image.Pt((imgRef.Max.X)/2 - (imgLic.Max.X)/2, (imgRef.Max.Y)/2 - (imgLic.Max.X)/2)
        imgTmp := image.NewRGBA(imgRef)
        draw.Draw(imgTmp, imgRef, picture1, image.ZP, draw.Src)
        draw.Draw(imgTmp, imgLic.Add(offset), picture2, image.ZP, draw.Over)
 
        imgResult ,err := os.Create("UnicornResult.jpg")
        if err != nil {
          log.Fatalf("failed to create: %s", err)
        }

        jpeg.Encode(imgResult, imgTmp, &jpeg.Options{jpeg.DefaultQuality})  
        defer imgResult.Close()
}


















