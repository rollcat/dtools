package dmenu

func assert(ex ...error) {
	for _, e := range ex {
		if e != nil {
			panic(e)
		}
	}
}
