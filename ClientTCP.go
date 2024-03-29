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
	//Commande du Client à affecter à l'image
	ORDER string
	//Image de type image.RGBA evnoyé par le client
  IMG *image.RGBA
}

func main() {

		//On vérifie que l'utilisateur entre bien le bon nombre d'arguments
		if len(os.Args) != 4 {
			fmt.Println("Il faut les arguments adresse:port chemin/vers/le/fichier grayscale|transparent|unicorn")
			return
		}
		if(os.Args[3] != "grayscale" && os.Args[3] != "transparent" && os.Args[3] != "unicorn"){
			fmt.Println("Seul grayscale, transparent et unicorn sont valide")
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
						os.Exit(1)
		    }

			fmt.Println("Traitement de l'image pour l'envoi...")
			//On crée un objet image.RGBA à partir de l'objet image.Image
			bounds := imgSrc.Bounds()
			imageToSend := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
			draw.Draw(imageToSend, imageToSend.Bounds(), imgSrc, bounds.Min, draw.Src)

			//On stocke l'objet de type image.RGBA que l'on souhaite envoyé dans une nouvelle structure de type imgStruct
			imageStruct := imgStruct{ORDER: os.Args[3], IMG: imageToSend}

			//On crée un objet de type Encoder pour envoyer la structure, le flux de sortie sera "connection"
			gobEncoder := gob.NewEncoder(connection)
			//On encode et envoi la structure
			gobEncoder.Encode(imageStruct)
			fmt.Println("Image envoyée")


}

func receiveFileFromServer (connection net.Conn) {

	//On crée une stucture temporaire vide de type imgStruct
	tmpstruct := &imgStruct{}

	//On crée un objet de type Decoder pour recevoir la structure, le flux d'entrée sera "connection"
	gobDecoder := gob.NewDecoder(connection)

	//On récupère et décode la structure recu
	gobDecoder.Decode(tmpstruct)

	// On crée un nouveau fichier ayant comme nom "gray" suivi du nom du fichier entré au début
	newFileName := os.Args[3] + "_" + filepath.Base(os.Args[2]) + ".png"
  	newfile, err := os.Create(newFileName)
  	if err != nil {
  	    panic(err.Error())
  	}
 	defer newfile.Close()

	// On encode en PNG l'objet image.RGBA dans le fichier créé
  	png.Encode(newfile,tmpstruct.IMG)

	fmt.Println("Image reçue")


}
