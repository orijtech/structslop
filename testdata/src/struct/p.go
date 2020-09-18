package p

type s struct{}

type s1 struct {
	i int
}

type s2 struct { // want "not implemented"
	i int
	j int
}
