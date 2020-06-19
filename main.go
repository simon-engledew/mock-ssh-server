package main // import "github.com/simon-engledew/mock-ssh-server"
import (
	"fmt"
	"github.com/gliderlabs/ssh"
	"github.com/simon-engledew/mock-ssh-server/pkg"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"io"
	"log"
	"os"
)

type Script func(w io.Writer, r io.Reader)

func Actor(script Script, enc encoding.Encoding) ssh.Handler {
	return func(s ssh.Session) {
		log.Print("client connected: ", s.RemoteAddr())
		script(enc.NewEncoder().Writer(s), enc.NewDecoder().Reader(s))
	}
}

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Fatal("usage: gosshim <script.star>")
	}

	script := os.Args[1]

	ssh.Handle(Actor(func(w io.Writer, r io.Reader) {
		err := pkg.RunScript(script, w, r)
		if err != nil {
			fmt.Fprint(w, err.Error()+"\r\n")
			log.Print(err)
		}
	}, unicode.UTF8))

	log.Fatal(ssh.ListenAndServe(":2222", nil))
}
