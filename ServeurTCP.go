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
  //Commande du Client à affecter à l'image
  ORDER string
  //Image de type image.RGBA evnoyé par le client
  IMG *image.RGBA
}

//création d'une structure de type job pour transmettre les instructions aux goRoutines
type job struct{
  x int
  h int
  imgsrc *image.RGBA
  imgdst *image.RGBA
  wg *sync.WaitGroup
}

// Constante pour le nombre de Goroutine disponible
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
        for {
      		connection, err := l.Accept()
      		if err != nil {
      			fmt.Println("Error: ", err)
            continue
      		}
      		fmt.Println("Client Connecté")
          //On lance une goroutine pour intéragir avec le client
      		go InteractClient(connection)
      	}

}

func InteractClient(connection net.Conn) {

    inputChannel := make(chan job, GoRoutinesNbr)

    //On crée une stucture temporaire vide de type imgStruct
    tmpstruct := &imgStruct{}

    //On crée un objet de type Decoder pour recevoir la structure, le flux d'entrée sera "connection"
    gobobj := gob.NewDecoder(connection)

    //On récupère et décode la structure recu
    gobobj.Decode(tmpstruct)

    fmt.Println("Image recu...")

    //on vérifie si tmpstruct.IMG à bien été remplit
    if tmpstruct.IMG == nil {
      fmt.Printf("Erreur image reçue")
      return
    }

    //On crée un objet image.RGBA vierge
    bounds := tmpstruct.IMG.Bounds()
    w, h := bounds.Max.X, bounds.Max.Y
    grayScale := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{w, h}})

    //Si on a recu 'unicorn' on souhaite dessiner la licorne au centre de l'image cible
    if(tmpstruct.ORDER == "unicorn"){
      draw_unicorn(grayScale)
    }

    fmt.Println("Traitement de l'image")


    start := time.Now()
    var wg sync.WaitGroup

    //On lance des goRoutines sur l'enssemble de la largeur de l'image selon la demande du Client
    //Transformation en noir et blanc
    if(tmpstruct.ORDER == "grayscale"){
      for channum := 0; channum < GoRoutinesNbr; channum++{
          go TransformToGrey(inputChannel)
      }
    }
    //Blanc vers alpha
    if(tmpstruct.ORDER == "transparent"){
      for channum := 0; channum < GoRoutinesNbr; channum++{
          go TransformTransparent(inputChannel)
      }
    }
    //Filigrane sur l'image cible
    if(tmpstruct.ORDER == "unicorn"){
      for channum := 0; channum < GoRoutinesNbr; channum++{
          go TransformFiligrane(inputChannel)
      }
    }


    wg.Add(1)

    //Lance une GoRoutine qui donnera les instruction à travers le channel 'inputChannel'
    go giveJob(inputChannel, &wg, tmpstruct.IMG, grayScale, w, h)

    wg.Wait()

    fmt.Println("Wait")

    elapsed := time.Since(start)
    log.Printf("Traitement took %s", elapsed)

    //On stocke l'objet de type image.RGBA que l'on souhaite envoyé dans une nouvelle structure de type imgStruct
    imageStruct := imgStruct{IMG: grayScale}

    //On crée un objet de type Encoder pour envoyer la structure, le flux de sortie sera "connection"
    gobEncoder := gob.NewEncoder(connection)
    //On encode et envoi la structure
    gobEncoder.Encode(imageStruct)

    fmt.Println("Image envoyée")

}

func TransformToGrey(jobChannel chan job){
  for{

      //Reception d'une structure du jobChannel
      job := <- jobChannel
      //Ajout 1 au WaitGroup
      job.wg.Add(1)

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

      //Reception d'une structure du jobChannel
      job := <- jobChannel
      //Ajout 1 au WaitGroup
      job.wg.Add(1)

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

      for y := 0; y < job.h; y++ {

          //On récupere l'objet de type color.Color correspondant aux coordonnés x,y de l'image
          imageColorSrc := job.imgsrc.At(job.x, y)
          imageColorDst := job.imgdst.At(job.x, y)
          //On récupère les valeur Red Blue Green correspondant
          redSrc, greenSrc, blueSrc, a := imageColorSrc.RGBA()
          redDst, greenDst, blueDst, _ := imageColorDst.RGBA()
          //On fait la moyenne de ces couleurs que l'on stocke dans mediumColor
        	col := color.RGBA{uint8((redSrc*(redDst))/65535>>8), uint8((greenSrc*(greenDst))/65535>>8), uint8((blueSrc*(blueDst))/65535>>8), uint8(a>>8)}
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

        //Ouverture de l'image que l'on souhaite avoir en filigrane
        image2, err1 := os.Open("licorne1.png")
        if err1 != nil {
          fmt.Println("failed to open: %s", err1)
          return
        }
        picture2, err := png.Decode(image2)
        if err != nil {
          fmt.Println("failed to decode: %s", err)
          return
        }
        defer image2.Close()

        imgRef := picture1.Bounds()
        imgLic := picture2.Bounds()
        //On crée l'offset, le centre de l'image de destination
        offset := image.Pt((imgRef.Max.X)/2 - (imgLic.Max.X)/2, (imgRef.Max.Y)/2 - (imgLic.Max.X)/2)

        //On remplit l'image de destination en blanc
        draw.Draw(picture1, picture1.Bounds(), &image.Uniform{color.White}, image.ZP, draw.Src)
        //On dessine une image au centre de l'image de destination
        draw.Draw(picture1, imgRef, picture1, image.ZP, draw.Src)
        draw.Draw(picture1, imgLic.Add(offset), picture2, image.ZP, draw.Over)

}
