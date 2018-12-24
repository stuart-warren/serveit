package access

type Permitted struct {
	methods []string
	users   []string
	groups  []string
}

func BlankPermit() Permitted {
	return Permitted{
		methods: []string{},
		users:   []string{},
		groups:  []string{},
	}
}

func (p Permitted) Methods() []string {
	return p.methods
}

func (p Permitted) Users() []string {
	return p.users
}

func (p Permitted) Groups() []string {
	return p.groups
}

func (p Permitted) MethodRO() Permitted {
	p.methods = []string{"GET"}
	return p
}

func (p Permitted) MethodRW() Permitted {
	p.methods = []string{"GET", "PUT", "POST", "PATCH"}
	return p
}

func (p Permitted) AllowMethods(methods ...string) Permitted {
	p.methods = append(p.methods, methods...)
	return p
}

func (p Permitted) AllowUsers(users ...string) Permitted {
	p.users = append(p.users, users...)
	return p
}

func (p Permitted) AllowGroups(groups ...string) Permitted {
	p.groups = append(p.groups, groups...)
	return p
}
