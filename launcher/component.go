package launcher

import "os/exec"

type Component struct {
	Name string
	cmd  *exec.Cmd
}

func NewComponent(name string) *Component {
	component := Component{Name: name}
	return &component
}

/*
typedef struct zl_comp_t {

	char name[32];
	char bin[_POSIX_PATH_MAX + 1];
	pid_t pid;
	int output;

	bool clean_stop;
	int fail_cnt;
	time_t start_time;

	enum {
	  ZL_COMP_AS_SHARE_NO,
	  ZL_COMP_AS_SHARE_YES,
	  ZL_COMP_AS_SHARE_MUST,
	} share_as;

	pthread_t comm_thid;

	zl_int_array_t restart_intervals;
	int min_uptime; // secs

  } zl_comp_t;
*/
