package main

import (
	tfea "github.com/christopherfriedrich/tf-ea/cmd/tf-ea"
	"github.com/christopherfriedrich/tf-ea/internal/log"
)

func main() {
	log.InitLogger()

	tfea.Execute()
}
