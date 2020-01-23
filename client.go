package main

import (
	"flag"
	"github.com/wo0lien/client/pool"
	"os"
)

const bufferSize = 1024

func main() {

	// Handle host:port, image and filter with flags
	host := flag.String("host", "127.0.0.1", "Nom d'hote du serveur")
	port := flag.Int("port", 8080, "Port du serveur")
	filter := flag.Int("filter", 1, "Filtre à utiliser : 1 = negatif,2 = greyscale, 3 = edge, 4 = median noise filter, 5 = mean noise filter")
	filePath := flag.String("path", "", "--REQUIRED-- Chemin relatif vers l'image")
	nb := flag.Int("Nb", 1, "Number of concurrent client laucnhed")

	flag.Parse()

	//exit program with log if no filePath is given
	if *filePath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	pool.Pool(*nb, *filter, *port, *host, *filePath)

}
