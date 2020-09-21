package p

type s struct{}

type s1 struct {
	i int
}

type s2 struct {
	i int
	j int
}

type s3 struct { // want "struct{x uint32; y uint64; z uint32} has size 24, could be 16, rearrange to struct{y uint64; x uint32; z uint32} for optimal size"
	x uint32
	y uint64
	z uint32
}
