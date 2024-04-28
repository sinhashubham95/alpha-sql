package pool

func removeFromConnections(v *[]*Connection, c *Connection) {
	for i, e := range *v {
		if e == c {
			lastIdx := len(*v) - 1
			(*v)[i] = (*v)[lastIdx]
			(*v)[lastIdx] = nil // Avoid memory leak
			*v = (*v)[:lastIdx]
			return
		}
	}
}
