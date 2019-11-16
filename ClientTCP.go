package main

import (
	"fmt"
	"net"
	"os"
	"encoding/gob"
	"image"
	"image/draw"
	"image/png"
	"path/filepath"
	_ "image/jpeg"
	_ "image/gif"
)


//création d'une structure de type imgStruct
type imgStruct struct{
  IMG *image.RGBA
}

func main() {

		//On vérifie que l'utilisateur entre bien le bon nombre d'arguments
		arguments := os.Args
		if len(arguments) != 3 {
			fmt.Println("Il faut les arguments adresse:port et chemin vers le fichier")
			return
		}

		//On initie la connexion à l'adresse et au port entré
		CONNECT := os.Args[1]
    	connection, err := net.Dial("tcp", CONNECT)
    	if err != nil {
        	fmt.Println("Error: ", err)
        	os.Exit(1)
    	}
    	fmt.Println("Connexion")

		//On lance les fonctions pour envoyer l'image et réceptionner l'image modifié
		sendFileToServer(connection)
		receiveFileFromServer(connection)


}



func sendFileToServer(connection net.Conn) {

			//On ouvre le fichier
			file, err := os.Open(os.Args[2])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer file.Close()

			//On recupère l'objet de type image.Image
			imgSrc, _, err := image.Decode(file)
			if err != nil {
	        	panic(err.Error())
		    }

			fmt.Println("Traitement de l'image pour l'envoi...")
			//On crée un objet image.RGBA à partir de l'objet image.Image
			bounds := imgSrc.Bounds()
			imageToSend := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
			draw.Draw(imageToSend, imageToSend.Bounds(), imgSrc, bounds.Min, draw.Src)

			//On stocke l'objet de type image.RGBA que l'on souhaite envoyé dans une nouvelle structure de type imgStruct
			imageStruct := imgStruct{IMG: imageToSend}

			//On crée un objet de type Encoder pour envoyer la structure, le flux de sortie sera "connection"
			gobEncoder := gob.NewEncoder(connection)
			//On encode et envoi la structure
			gobEncoder.Encode(imageStruct)
			fmt.Println("Image envoyer")


}

func receiveFileFromServer (connection net.Conn) {

	//On crée une stucture temporaire vide de type imgStruct
	tmpstruct := &imgStruct{}

	//On crée un objet de type Decoder pour recevoir la structure, le flux d'entrée sera "connection"
	gobDecoder := gob.NewDecoder(connection)

	//On récupère et décode la structure recu
	gobDecoder.Decode(tmpstruct)

	// On crée un nouveau fichier ayant comme nom "gray" suivi du nom du fichier entré au début
	newFileName := "gray" + filepath.Base(os.Args[2]) + ".png"
  	newfile, err := os.Create(newFileName)
  	if err != nil {
  	    panic(err.Error())
  	}
 	defer newfile.Close()

	// On encode en PNG l'objet image.RGBA dans le fichier créé
  	png.Encode(newfile,tmpstruct.IMG)

	fmt.Println("Image recu")


}
