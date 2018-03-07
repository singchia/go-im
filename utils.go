package main

func validate(uid string) bool {
	if uid == "" || uid[0] < 'a' || uid[0] > 'z' {
		return false
	}
	for _, r := range uid[1:len(uid)] {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '1' || r > '9') {
			return false
		}
	}
	return true
}
