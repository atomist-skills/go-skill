package legacy

const (
	GetUserQueryName = "get-user"
)

type DockerImageUser struct {
	ImageUser string `edn:"docker.image/user,omitempty"`
}

func MockGetUserForLocalEval(user string) DockerImageUser {
	return DockerImageUser{
		ImageUser: user,
	}
}
