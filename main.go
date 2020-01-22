package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
)

func computePhash(img string, hashAlgorithm contrib.ImgHashBase) string {
	image := gocv.IMRead(img, gocv.IMReadColor)
	if image.Empty() {
		log.Printf("cannot read image %s\n", img)
		return ""
	}
	defer image.Close()
	hash := gocv.NewMat()
	defer hash.Close()
	hashAlgorithm.Compute(image, &hash)
	if hash.Empty() {
		log.Printf("error computing hash for %s\n", img)
		return ""
	}
	return fmt.Sprintf("%x", hash.ToBytes())

}
func main() {
	hashes := map[string]string{}
	hashAlgorithm := contrib.PHash{}
	extension := flag.String("extension", "", "file extension")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		log.Printf("you must specify at least one folder.")
		return
	}
	for _, folder := range args {
		files, err := ioutil.ReadDir(folder)
		if err != nil {
			log.Fatal(err)
		}

		for i, f := range files {
			if !f.IsDir() {
				if strings.HasSuffix(f.Name(), *extension) {
					log.Printf("[%d/%d] processing %s", i, len(files), f.Name())
					hash := computePhash(folder+f.Name(), hashAlgorithm)
					hashes[f.Name()] = hash
				}
			}
		}
	}
	jsonResult, err := json.MarshalIndent(hashes, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", jsonResult)
}
