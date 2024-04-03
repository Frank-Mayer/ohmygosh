package runtime

var BuiltinCommands map[string]func(*Command, *IoProvider) error

func init() {
	BuiltinCommands = map[string]func(*Command, *IoProvider) error{
		"cd":     execute_cd,
		"exit":   execute_exit,
		"echo":   execute_echo,
		"cat":    execute_cat,
		"export": execute_export,
		"unset":  execute_unset,
		"whoami": execute_whoami,
		"pwd":    execute_pwd,
		"which":  execute_which,
		"type":   execute_type,
		"sudo":   execute_sudo,
		"yes":    execute_yes,
		"true":   execute_true,
		"false":  execute_false,
		"sleep":  execute_sleep,
	}
}
