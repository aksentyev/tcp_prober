package util

func IncludesStr(source []string, target string) bool {
    for _, el := range source {
        if el == target {
            return true
        }
    }
    return false
}
