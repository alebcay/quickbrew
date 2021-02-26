package utils

import "fmt"

func StringSliceContains(v string, a []string) bool {
    for _, i := range a {
        if i == v {
			fmt.Printf("match\tfound %v\n", v)
            return true
        }
    }
    return false
}
