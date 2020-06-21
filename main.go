package main // import "github.com/simon-engledew/mock-ssh-server"
import (
	"fmt"
	"github.com/gliderlabs/ssh"
	"github.com/simon-engledew/mock-ssh-server/pkg"
	"golang.org/x/text/encoding/unicode"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Fatal("usage: gosshim <script.star>")
	}

	script, err := pkg.Load(os.Args[1], nil)
	if err != nil {
		log.Fatal(err)
	}

	encoding := unicode.UTF8

	ssh.Handle(func(s ssh.Session) {
		log.Print("client connected: ", s.RemoteAddr())

		w := encoding.NewEncoder().Writer(s)
		r := encoding.NewDecoder().Reader(s)

		err := script(w, r)
		if err != nil {
			if _, fmtErr := fmt.Fprint(w, err.Error()+"\r\n"); fmtErr != nil {
				log.Print(err)
			}
			log.Print(err)
		}
	})

	log.Fatal(ssh.ListenAndServe(":2222", nil))
}
