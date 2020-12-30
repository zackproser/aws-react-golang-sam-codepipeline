package main

//Takes a slice of hostname strings and returns
//a map denoting how many times each hostname occurs
func tallyCounts(s []string) map[string]int {
	//Map for holding counts of each item
	m := make(map[string]int)
	for _, v := range s {
		if _, present := m[v]; present {
			m[v] += 1
		} else {
			m[v] = 1
		}
	}
	return m
}
