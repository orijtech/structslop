package p

type s struct{}

type s1 struct {
	i int
}

type s2 struct {
	i int
	j int
}

type s3 struct { // want "struct{.+} has size 20, could be 16, rearrange to struct{y uint64; x uint32; z uint32} for optimal size"
	x uint32
	y uint64
	z uint32
}

type s4 struct { // want `struct{.+} has size 32, could be 20, rearrange to struct{_ \[0\]func\(\); i1 int; i2 int; a3 \[3\]bool; b bool} for optimal size`
	b  bool
	i1 int
	i2 int
	a3 [3]bool
	_  [0]func()
}
