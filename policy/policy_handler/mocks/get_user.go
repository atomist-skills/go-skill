package mocks

const (
	GetUserQueryName = "get-user"
)

type DockerImageUser struct {
	ImageUser string `edn:"docker.image/user,omitempty"`
}

func MockGetUser(user string) DockerImageUser {
	return DockerImageUser{
		ImageUser: user,
	}
}
