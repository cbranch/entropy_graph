package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
)

type entropy_blocks struct {
	entropy  float64
	block    []byte
	off_from int
	off_to   int
}

func main() {

	var block_size int
	var suspecious_entropy float64
	var filename string
	var output string

	if len(os.Args) < 2 {
		fmt.Printf("Usage : %s <filename> <block_size> <suspecious_entropy> <output>\n", os.Args[0])
		return
	}

	filename = os.Args[1]

	if len(os.Args) > 2 {
		temp, err := strconv.Atoi(os.Args[2])
		if err != nil || temp <= 0 {
			fmt.Println("Invalid block_size")
			return
		}
		block_size = temp
	} else {
		block_size = 32
	}

	if len(os.Args) > 3 {
		temp, err := strconv.ParseFloat(os.Args[3], 64)
		if err != nil || temp > 8 {
			fmt.Println("Invalid suspecious entropy")
			return
		}
		suspecious_entropy = temp
	} else {
		suspecious_entropy = 5.0
	}

	if len(os.Args) > 4 {
		output = os.Args[4]
	} else {
		output = "point.png"
	}

	fmt.Printf("[+] Filename %s, block size : %d, suspecious entropy %g, output %s\n",
		filename, block_size, suspecious_entropy, output)

	brange := make([]byte, 256)
	for i := range brange {
		brange[i] = byte(i)
	}

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("[*] Total Entropy %g\n", H(data, brange))

	ent_blocks := generate_entropy_blocks(data, block_size, brange)

	// Graph
	fmt.Println("[+] Graphing..")
	if err := graph(ent_blocks, suspecious_entropy, output); err != true {
		fmt.Println("[-] Unble to graph")
	}

	// Suspecious blocks
	n_suspecious := 0

	for i := range ent_blocks {
		if ent_blocks[i].entropy >= suspecious_entropy {
			n_suspecious++
		}
	}

	if n_suspecious > 0 {
		fmt.Printf("[*] Suspecious blocks : %d\n", n_suspecious)
		for i := range ent_blocks {
			if ent_blocks[i].entropy >= suspecious_entropy {
				fmt.Printf("[0x%.8x,0x%.8x] -> %g\n", ent_blocks[i].off_from,
					ent_blocks[i].off_to,
					ent_blocks[i].entropy)
			}
		}
	} else {
		fmt.Println("No suspecious blocks\n")
	}

}

func graph(ent_blocks []entropy_blocks, suspecious_entropy float64, output string) bool {

	p, err := plot.New()
	if err != nil {
		return false
	}

	p.Title.Text = "Entropy"
	p.X.Label.Text = "offsets"
	p.Y.Label.Text = "Entropy"

	err = plotutil.AddLinePoints(p,
		"line", graph_xy_entropy(ent_blocks),
		"suspecious", graph_xy_suspecious(suspecious_entropy, ent_blocks))
	if err != nil {
		return false
	}

	if err := p.Save(20, 20, output); err != nil {
		return false
	}

	return true
}

func graph_xy_suspecious(suspecious_entropy float64, ent_block []entropy_blocks) plotter.XYs {

	pts := make(plotter.XYs, len(ent_block))

	for i := range pts {
		pts[i].X = float64(ent_block[i].off_from)
		pts[i].Y = suspecious_entropy
	}

	return pts
}

func graph_xy_entropy(ent_blocks []entropy_blocks) plotter.XYs {

	pts := make(plotter.XYs, len(ent_blocks))

	for i := range pts {
		pts[i].X = float64(ent_blocks[i].off_from)
		pts[i].Y = float64(ent_blocks[i].entropy)
	}

	return pts
}

func generate_entropy_blocks(data []uint8, block_size int, brange []byte) []entropy_blocks {

	n_blocks := len(data)/(block_size) + 1
	ent_blocks := make([]entropy_blocks, n_blocks)

	// fill entropy blocks with data as blocks
	cur_block := 0
	from := 0
	overflow := 0
	for from = 0; from < len(data); from += block_size {

		// don't overflow
		if from+block_size > len(data) {
			overflow = ((from + block_size) % len(data))
		}

		to := from + block_size - overflow
		ent_blocks[cur_block].block = data[from:to]
		ent_blocks[cur_block].off_from = from
		ent_blocks[cur_block].off_to = to
		ent_blocks[cur_block].entropy = H(ent_blocks[cur_block].block, brange)

		cur_block++

		// end
		if from >= len(data) || cur_block >= (len(data)/(block_size))+1 {
			break
		}

	}

	return ent_blocks
}

func H(data []uint8, brange []byte) float64 {
	var p_x float64
	var entropy = float64(0)

	for i := range brange {
		p_x = float64(CountBytes(brange[i], Uint8ToBytes(data))) / float64(len(data))
		if p_x > float64(0) {
			entropy += -p_x * math.Log2(p_x)
		}
	}

	return entropy
}

func CountBytes(needle byte, haystack []byte) int {
	var count = 0

	for i := range haystack {
		if needle == haystack[i] {
			count++
		}
	}

	return count
}

func Uint8ToBytes(a []uint8) []byte {
	b := make([]byte, len(a))

	for i := range a {
		b[i] = byte(a[i])
	}

	return b
}
