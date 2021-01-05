package main

type ConstantIP4AM struct {
	ip int64
}

func NewConstantIP4AM(ip int64) *ConstantIP4AM {
	return &ConstantIP4AM{ip: ip}
}

func (c *ConstantIP4AM) Assign() int64 {
	return c.ip
}

func (c *ConstantIP4AM) AssignSpecificIP(ip uint32) bool {
	return int64(ip) == c.ip
}

func (c *ConstantIP4AM) Release(ip uint32) {
}

func (c *ConstantIP4AM) IsIPOutOfRange(ip uint32) bool {
	return int64(ip) != c.ip
}

func (c *ConstantIP4AM) IsIPInRange(ip uint32) bool {
	return int64(ip) == c.ip
}

func (c *ConstantIP4AM) GetFirst() uint32 {
	return uint32(c.ip)
}

func (c *ConstantIP4AM) GetLast() uint32 {
	return uint32(c.ip)
}

func (c *ConstantIP4AM) Count() uint32 {
	return 1
}

func (c *ConstantIP4AM) Close() error {
	return nil
}
