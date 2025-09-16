package ports

type NicknameGenerator interface {
	Generate() string
	Release(name string)
}
