# Repeat

Part of BAD2BEEF's toolbox

## Overview

`repeat` commands over inventory data stored in `.csv` or `.json` files. Powershell does an excellent job at this. xargs can get the job done as well. `parallel`, `ppss`, and `psexec` are all things that exist. `repeat` allows for use of a single tool with consistent behavior across multiple platforms.

Inventory values can be substituted into each repeat via two methods: variable substitution and environment variables. Command and arguments are checked for variable substitutions in the form of `${VARIABLE}` where `VARIABLE` is a column or property name from inventory. Additionally, environment variables are set for each column or property name before execution.

Two to four log entries will be written to stdout for each repeated command. Each line is prefixed by the current date, time, a random 32 bit integer in hex format for correlation of nodes across multiple lines, and an indicator character unique to each log entry type. The first and last lines show the start and end of the repeated command and are respectively indicated by a `>` and `<` . End lines also include the repeated commands' exit code. Command `stdout` and `stderr` are written out in-between indicated by `1` and `2`, respectively. `stdout` and `stderr` may span multiple lines but are only prefixed by date and time stamps, and output indicators once.

## Building

    make

or

    go build repeat.go repeat-CSV.go repeat-Filter.go repeat-JSON.go repeat-Node.go

## Usage

    repeat [-async] [-inventory [inventory/|inventory.[csv|json]]] [-bash|-cmd|-ps|-pwsh] [Key[==|!=|~=|<=|>=]Value,...] - command [argument,...]

### Options

- *-async* Run asynchronously
- *-inventory* Specify inventory file or directory location
- *-bash|-cmd|-ps|-pwsh* Prefix command with one-shot helpers for common shells
- *Key?=Value* *?:=, !, ~, <, >* Specify a filter for inventory items
- *-* Signify end of options, remaining items are the command and arguments
- *command, argument* Command and arguments to repeat

### Examples

#### Multi-Filter Echo Example

    > .\repeat.exe -inventory .\sample-inv\ -cmd timezone~=America/ department==Training - 'echo ${node} ${owner} %city%'
    ...
    2021/03/21 13:19:34.739137 CEC86E7B > map[address:1.1.1.1 city:Kansas City department:Training node:clibreyls-laptop owner:clibreyls timezone:America/Chicago type:laptop]
    2021/03/21 13:19:34.766890 CEC86E7B 1 clibreyls-laptop clibreyls Kansas City
    2021/03/21 13:19:34.766890 CEC86E7B < 0
    2021/03/21 13:19:34.767964 E807D54E > map[address:1.1.1.2 city:Riachão das Neves department:Training node:cmossnf-mac owner:cmossnf timezone:America/Bahia type:mac]
    2021/03/21 13:19:34.794283 E807D54E 1 cmossnf-mac cmossnf Riachao das Neves
    2021/03/21 13:19:34.794543 E807D54E < 0
    2021/03/21 13:19:34.795926 621A7080 > map[address:1.1.1.3 city:Tacoma department:Training node:jvancasselpp-mac owner:jvancasselpp timezone:America/Los_Angeles type:mac]
    2021/03/21 13:19:34.836047 621A7080 1 jvancasselpp-mac jvancasselpp Tacoma
    2021/03/21 13:19:34.836047 621A7080 < 0
    ...
    >
