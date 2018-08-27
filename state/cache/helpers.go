package cache

func CheckCacheForEvent(value interface{}) {
	//figure out which cache to check
	switch v := context.(type) {
	case *apievent:

	case apievent:

	case dmpsevent:

	case *dmpsevent:

	default:
	}

}
