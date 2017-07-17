package main

import (
	"fmt"
	"io"
	"net/http"

	u "github.com/ipfs/go-ipfs-util"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreunix"
	"github.com/ipfs/go-ipfs/path"
	"github.com/ipfs/go-ipfs/repo/fsrepo"
	"golang.org/x/net/context"
)

var gnode *core.IpfsNode

func ServeIpfsRand(w http.ResponseWriter, r *http.Request) {
	read := io.LimitReader(u.NewTimeSeededRand(), 2048)

	str, err := coreunix.Add(gnode, read)
	if err != nil {
		w.WriteHeader(504)
		w.Write([]byte(err.Error()))
	} else {
		w.Write([]byte(str))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	name := "foo" // assumes you've created a key w/ this name using cil
	pk, err := gnode.GetKey(name)
	if err != nil {
		panic(err)
	}
	fmt.Println("publishing", str, "to", name)
	// We hang at this line
	err = gnode.Namesys.Publish(ctx, pk, path.Path(str))
	if err != nil {
		panic(err)
	}
	fmt.Println("done")
}

func main() {
	r, err := fsrepo.Open("~/.ipfs")
	if err != nil {
		panic(err)
	}

	// Make our 'master' context and defer cancelling it
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := &core.BuildCfg{
		Repo:   r,
		Online: true,
	}

	node, err := core.NewNode(ctx, cfg)
	if err != nil {
		panic(err)
	}

	// Set the global node for access in the handler
	gnode = node

	http.HandleFunc("/ipfsobject", ServeIpfsRand)
	http.ListenAndServe(":8080", nil)
}
