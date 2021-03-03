package db

type UnionFind struct {
	Parent []int64
	Size []int64
	Count int64
}

func NewUnionFind(size int64) *UnionFind {
	res := &UnionFind{
		Parent: make([]int64, size),
		Size:   make([]int64, size),
		Count:  size,
	}
	for i := int64(0); i < size; i++ {
		res.Parent[i] = i
		res.Size[i] = 1
	}
	return res
}

func (u *UnionFind) Find(p int64) int64 {
	for p != u.Parent[p] {
		p = u.Parent[p]
	}
	return p
}

func (u *UnionFind) Connected(p, q int64) bool {
	return u.Find(p) == u.Find(q)
}

func (u *UnionFind) Union(p, q int64) {
	pRoot := u.Find(p)
	qRoot := u.Find(q)

	if pRoot == qRoot {
		return
	}

	if u.Size[pRoot] < u.Size[qRoot] {
		u.Parent[pRoot] = qRoot
		u.Size[qRoot] += u.Size[pRoot]
	} else {
		u.Parent[qRoot] = pRoot
		u.Size[pRoot] += u.Size[qRoot]
	}
}

func (u *UnionFind) CountSets() []int64 {
	parent := make([]int64, 0, 2)
	for i := int64(0); i < u.Count; i++ {
		if u.Find(i) == i {
			parent = append(parent, i)
		}
	}
	return parent
}

