# fetter

[![Test status](https://github.com/juan-leon/fetter/actions/workflows/test.yml/badge.svg)](https://github.com/juan-leon/fetter/actions/fetter/test.yml)
[![Lint status](https://github.com/juan-leon/fetter/actions/workflows/lint.yml/badge.svg)](https://github.com/juan-leon/fetter/actions/fetter/lint.yml)
[![Release](https://img.shields.io/github/release/juan-leon/fetter.svg)](https://github.com/juan-leon/fetter/releases/latest)
[![codecov](https://codecov.io/gh/juan-leon/fetter/branch/main/graph/badge.svg?token=8VJ64CHLMZ)](https://codecov.io/gh/juan-leon/fetter)
[![Go Report Card](https://goreportcard.com/badge/github.com/juan-leon/fetter)](https://goreportcard.com/report/github.com/juan-leon/fetter)

Move processes into control groups, and trigger other actions, based on configurable events.

Fetter can detect when new processes are launched and, based on the executable
name, move those project into pre-defined process control groups.  The control
groups can be configured with limits: CPU, RAM, and number of (children)
processes they allow..  This makes fetter a way of protecting your system from
an application that (because of its design or because a bug) misbehave and
request too many resources.

See a [sample configuration] to grasp the possibilities that fetter has.

## Uses

### Controlling the maximum uses of resources an application can use.

The goal would be to avoid an application from using too many resources and
leaving the rest of the system (or just the window manager) unusable.

Here is a quick configuration example:

```yaml
rules:
  compilation:
    paths: [/usr/bin/make]
    action: execute
    group: compilation

# Control group definition
groups:
  compilation:
    # We do not allow compilations to use more than 90% of CPU horse power, to
    # avoid our code editor to have "lag".
    cpu: 90
    # Max megabytes that we allow the sum of all compilation processes to use.
    ram: 4500
```

In this example, `make` will probably spawn other processes (gcc, go, rustc,
etc.).  But since fetter receives an event as soon as `make` is executed, those
processes will be in same control group and their combined usage cannot go over
the limits you set up.

You can apply that same principle to any other application (see [sample
configuration]).  If you suspect that your browser uses a video-conference plugin
that sometimes freezes your whole computer, fetter can be the solution: it is
better if just the video conference tab dies because of lack of RAM.

Note that even if you may use fetter with services defined at OS level (like a
webserver or database), typically the service manager (like systemd, for
instance) already have a nice way of doing that (configuration droplets, for
systemd) that take advantage of service manager use of control groups.  I would
recommend using those mechanisms.  However, fetter would be the way for setting
limits to those applications launched via user actions (or crons).  Same goes
for containers: docker and container orchestration tools like kubernetes already
have their own and idiomatic way to declare limits (that also make use of
control groups).

Fetter can detect also when files are read or written, so you can configure
moving a process into a control group when they do that (actions `read`, `write`
and `execute` are supported).  There is more info in the configuration example.

Currently fetter allows to define three kind of limits for each control group.
By default all of three are unlimited.

* **ram**: Max amount of RAM (in megabytes) that the aggregation of processes in
  that group can use.

* **cpu**: Max amount of CPU (in %-age over the total CPU power of all the CPU
  cores of the system) the aggregation of processes in that group can use.

* **pids**: Max amount of processes the control group can contain.  Over that
  limit processes will not be able to fork new children (for instance, creating
  a new tab in a browser will yield a error page).

### Triggering actions when applications are started

Here is an example.

```yaml
rules:
  audit:
    paths: [/usr/bin/sudo]
    action: execute
    # Whenever sudo is executed (from a shell or any other process), trigger
    # 'send-mail' (defined in trigger sections) will be executed.
    trigger: send-mail

triggers:
  send-mail:
    # This is the tool that will be executed, in background, when rule is
    # triggered.  Results will be logged.
    run: /opt/utils/send-mail.sh
    # These are the arguments for tools above.  The tool can find additional
    # info about the process that triggered the action by looking at the
    # environment variables the type 'FETTER_X'.  For example, FETTER_PID will
    # have the pid of the process that triggered the action.  See docs to see
    # the values of X supported.
    args: ['audit@mycompany.org', 'cto@mycompany.org']
    # By default, the process defined by trigger will be ran as user nobody,
    # since it has little privileges.  If you want to use any other user,
    # including root, this would be the place to declare that.
    user: nobody
```

Of course, for the case of `sudo` already are other alternatives for usage
audit.  The key here is that you can trigger any action you like (as long as you
want to write the logic) as soon some command you are interested in is
executed (as opposed to waiting for the command to end).

As with control groups, you can define triggers for scenarios where some
specific files are written or read.  While that leads to interesting
possibilities, that is nothing that cannot be done with incron or some other
thing based on inotify.  The advantage of using about fetter is that there are
no inotify events for execution (the `in_access` event is used for reading in
general).

The process executed by the trigger will receive the arguments you specify and
will have available some environment variables with relevant information.  Those
are the more useful variables (other variables are low level details of syscall;
you can see the full set by printing environment from trigger)

| Variable       | Value                                                   |
|----------------|---------------------------------------------------------|
| FETTER_PID     | Pid of the new process (or process doing the read/write |
| FETTER_PPID    | Pid of its parent                                       |
| FETTER_SYSCALL | This would be execve when trigger action is "execute"   |
| FETTER_RESULT  | Result of syscall                                       |
| FETTER_EXE     | Executable file                                         |
| FETTER_UID     | UID of process                                          |
| FETTER_EUID    | Effective UID of process                                |
| FETTER_TTY     | (only if process was spawned from a tty)                |

### Freezing processes to examine them

You can define the freeze property of a control group to true to freeze
processes that do something, so you can examine them: their state, parent,
memory, etc.

For instance, if you are curious about what process is modifying some file, when
and why.  Or if you want to know who tries to peek to a specific file:

```yaml
rules:
  # This is an example where a process reading a file will be frozen in place by
  # the operating system
  freezes:
    paths: [/root/my-secret-password]
    action: read
    group: honeypot

groups:
  honeypot:
    freeze: true
```

A frozen process will not continue its execution, and it cannot be killed,
unless removed from cgroup, or cgroup is manually thawed.

### Killing processes

Example:

```
rules:
  deaths:
    paths: [/my/forbidden/file]
    action: write
    trigger: KILL
```

KILL is not a real trigger, but a convention to say to fetter: kill whatever
process is doing that action as soon as possible.  In this case, whenever a
process writes to the file in path, process will be instantly killed.

## Usage

After writing the configuration file (comments in [sample configuration] work as
documentation), just type `fetter run`, as root.  You can add the flag
`--daemon` if you want the program to detach from the terminal.  You can wrap
the `fetter run` into a systemd (or equivalent) definition so its status is
managed automatically.

You can also use the flag `--scan` to scan already active processes and classify
those in control groups as per the rules.

Type `fetter --help` or `fetter CMD --help` to see other sub-commands and options.

```
Available Commands:
  clean       Delete fetter cgroups
  quick-run   Scan currently running processes according to rules and exit
  run         Listen for rules defined in configuration and act accordlingly

Flags:
  -c, --config string   Path to configuration file (default "/etc/fetter/config.yaml")
```

Note that fetter will write logs to the file specified in configuration (as well
to stderr, unless `--daemon` is used).

## How does it work?

Upon starting, fetter connects to Linux kernel, using `netlink_audit` family
over netlink protocol, and builds the needed auditing rules bases on the
configuration file.  It uses a multicast client, in order to play nice with
auditd daemons (if any).  After that is done, it listen for events and act
accordingly, using cgroups Linux API.

Triggers are launched in background threads, so that new events are processed
with no delay.

One advantage of using `netlink_audit` is that it should be pretty easy to
extend fetter to trigger some code on execution of arbitrary linux syscalls (for
example `chown` or `mount`).

Both the creation of an audit client and the ability to move processes to
control groups require root privileges.  Also, fetter only works with Linux; I
am not familiar with audit and control groups (or equivalent) on other OS's.


### Installing

#### Via release

Go to [releases page] and download the binary you want.  Decompress the file and
copy the binary to your path.

#### Via local compilation

```
$ git clone https://github.com/juan-leon/fetter
$ cd fetter
$ make
```

You can use `goreleaser build --single-target` instead of make.

### Contributing

Feedback, ideas and pull requests are welcomed.

[sample configuration]: examples/documented-example.yaml
[releases page]: https://github.com/juan-leon/fetter/releases
