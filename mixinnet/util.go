package mixinnet

import (
	"sort"
)

func HashMembers(ids []string) string {
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	var in string
	for _, id := range ids {
		in = in + id
	}
	return NewHash([]byte(in)).String()
}
