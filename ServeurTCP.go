package main

import (
  "fmt"
  "net"
  "os"
  "image"
  "image/png"
  "image/color"
  "encoding/gob"
  "time"
  "log"
  "sync"
  "image/draw"
)

//création d'une structure de type imgStruct
type imgStruct struct{
  ORDER string
  IMG *image.RGBA
}

type job struct{
  x int
  h int
  imgsrc *image.RGBA
  imgdst *image.RGBA
  wg *sync.WaitGroup
}

const GoRoutinesNbr int = 1000


func main() {

        ///On vérifie que l'utilisateur entre bien le bon nombre d'arguments
        arguments := os.Args
        if len(arguments) != 2 {
                fmt.Println("Il faut entrer le port d'écoute comme argument")
                return
        }

        ///On initialise le serveur avec le port entré
        PORT := ":" + arguments[1]
        l, err := net.Listen("tcp", PORT)
        if err != nil {
                fmt.Println(err)
                return
        }
        defer l.Close()

        //On crée une boucle infini
        numCustomer := 0
        for {
      		connection, err := l.Accept()
          numCustomer += 1
      		if err != nil {
      			fmt.Println("Error: ", err)
      			os.Exit(1)
      		}
      		fmt.Println("Client Connecté")
          //On lance une goroutine pour intéragir avec le client
      		go InteractClient(connection, numCustomer)
      	}

}

func InteractClient(connection net.Conn, numClient int) {

    inputChannel := make(chan job, GoRoutinesNbr)

    //On crée une stucture temporaire vide de type imgStruct
    tmpstruct := &imgStruct{}

    //On crée un objet de type Decoder pour recevoir la structure, le flux d'entrée sera "connection"
    gobobj := gob.NewDecoder(connection)

    //On récupère et décode la structure recu
    gobobj.Decode(tmpstruct)

    fmt.Println("Image recu du client numéro : ", numClient)

    //On crée un objet image.RGBA vierge
    bounds := tmpstruct.IMG.Bounds()
    w, h := bounds.Max.X, bounds.Max.Y
    grayScale := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{w, h}})

    if(tmpstruct.ORDER == "unicorn"){
      draw_unicorn(grayScale)
    }

    fmt.Println("Traitement de l'image")

    //On lance des goroutines "TransformToGrey" sur toute la largeur de l'image

    start := time.Now()
    var wg sync.WaitGroup

    if(tmpstruct.ORDER == "gray"){
      for channum := 0; channum < GoRoutinesNbr; channum++{
          go TransformToGrey(inputChannel)
      }
    }
    if(tmpstruct.ORDER == "transparent"){
      for channum := 0; channum < GoRoutinesNbr; channum++{
          go TransformTransparent(inputChannel)
      }
    }
    if(tmpstruct.ORDER == "unicorn"){
      for channum := 0; channum < GoRoutinesNbr; channum++{
          go TransformFiligrane(inputChannel)
      }
    }


    wg.Add(1)

    go giveJob(inputChannel, &wg, tmpstruct.IMG, grayScale, w, h)

    wg.Wait()

    fmt.Println("Wait")
    //wg.Wait()

    elapsed := time.Since(start)
    log.Printf("Traitement took %s", elapsed)
    //On stocke l'objet de type image.RGBA que l'on souhaite envoyé dans une nouvelle structure de type imgStruct
    imageStruct := imgStruct{IMG: grayScale}

    //On crée un objet de type Encoder pour envoyer la structure, le flux de sortie sera "connection"
    gobEncoder := gob.NewEncoder(connection)
    //On encode et envoi la structure
    gobEncoder.Encode(imageStruct)

    fmt.Println("Image envoyer")

}

func TransformToGrey(jobChannel chan job){
  for{

      job := <- jobChannel
      job.wg.Add(1)
      //fmt.Println("Ajout")

      for y := 0; y < job.h; y++ {

          //On récupere l'objet de type color.Color correspondant aux coordonnés x,y de l'image
          imageColor := job.imgsrc.At(job.x, y)
          //On récupère les valeur Red Blue Green correspondant
          red, green, blue, _ := imageColor.RGBA()
          //On fait la moyenne de ces couleurs que l'on stocke dans mediumColor
          mediumColor := uint16((red + green + blue) / 3)
          //On crée la nouvelle couleur à partir de la moyenne décalé de 8 bits vers la droite que l'on converti en uint8
          newColor := color.Gray{uint8(mediumColor >> 8)}
          //On écrit cette couleur dans l'image vierge aux coordonnés x y
          job.imgdst.Set(job.x, y, newColor)

        }
      job.wg.Done()
    }
}

func TransformTransparent(jobChannel chan job){
  for{

      job := <- jobChannel
      job.wg.Add(1)
      //fmt.Println("Ajout")

      for y := 0; y < job.h; y++ {

          //On récupere l'objet de type color.Color correspondant aux coordonnés x,y de l'image
          imageColor := job.imgsrc.At(job.x, y)
          //On récupère les valeur Red Blue Green correspondant
          red, green, blue, _ := imageColor.RGBA()
          //On fait la moyenne de ces couleurs que l'on stocke dans mediumColor
          if ((red == 65535) && (green == 65535) && (blue == 65535)) {
        	col := color.RGBA{uint8(red>>8), uint8(green>>8), uint8(blue>>8), 0}
        	job.imgdst.Set(job.x, y, col)
        	} else {
        		job.imgdst.Set(job.x, y, job.imgsrc.At(job.x, y))
        	}

        }
      job.wg.Done()
    }
}

func TransformFiligrane(jobChannel chan job){
  for{

      job := <- jobChannel
      job.wg.Add(1)
      //fmt.Println("Ajout")

      for y := 0; y < job.h; y++ {

          //On récupere l'objet de type color.Color correspondant aux coordonnés x,y de l'image
          imageColorSrc := job.imgsrc.At(job.x, y)
          imageColorDst := job.imgdst.At(job.x, y)
          //On récupère les valeur Red Blue Green correspondant
          redSrc, greenSrc, blueSrc, a := imageColorSrc.RGBA()
          redDst, greenDst, blueDst, _ := imageColorDst.RGBA()
          //On fait la moyenne de ces couleurs que l'on stocke dans mediumColor
        	col := color.RGBA{uint8((redSrc-redDst)>>8), uint8((greenSrc-greenDst)>>8), uint8((blueSrc-blueDst)>>8), uint8(a>>8)}
        	job.imgdst.Set(job.x, y, col)

        }
      job.wg.Done()
    }
}

func giveJob(jobChannel chan job, wg *sync.WaitGroup, imgsrc *image.RGBA, imgdst *image.RGBA, width int, height int){

  fmt.Println("GiveJob")

  for x := 0; x < width; x++ {
    toPush := job{x: x, h: height, imgsrc: imgsrc, imgdst: imgdst, wg: wg}
    jobChannel <- toPush
  }

  wg.Done()

}

func draw_unicorn(picture1 *image.RGBA) {

        image2, err1 := os.Open("licorne1.png")
        if err1 != nil {
          log.Fatalf("failed to open: %s", err1)
        }
        picture2, err := png.Decode(image2)
        if err != nil {
          log.Fatalf("failed to decode: %s", err)
        }
        defer image2.Close()

        imgRef := picture1.Bounds()
        imgLic := picture2.Bounds()
        offset := image.Pt((imgRef.Max.X)/2 - (imgLic.Max.X)/2, (imgRef.Max.Y)/2 - (imgLic.Max.X)/2)
        //imgTmp := image.NewRGBA(imgRef)
        draw.Draw(picture1, imgRef, picture1, image.ZP, draw.Src)
        draw.Draw(picture1, imgLic.Add(offset), picture2, image.ZP, draw.Over)

        imgResult ,err := os.Create("UnicornResult.png")
        if err != nil {
          log.Fatalf("failed to create: %s", err)
        }

        png.Encode(imgResult, picture1)
        defer imgResult.Close()
}
