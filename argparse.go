package cli

import "flag"

//ParseArgumentsInterspersed allows argument parsing to be more flexible than
//what is provided natively in the flag package.
//In the flag package, all flag options must be specified before any other arguments.
//Parsing stop when the first non-flag argument is reached.
//ParseArgumentsInterspersed allows non-flag arguments to be parsed within the
//normal listing of flag arguemtns.
//
//For example, with f having the -value and -count flags set, the following command
//arguments would set -value and -count on f and return "one" and "two" as the return
//variable params.
//	[]string{"one" "-count" "22" "two" "-value" "foobar"}
//Essentially, the order of the arguments do not matter when parsing. And non-flag
//arguments are returned in params.
//Err will be any error returned from flag.FlagSet.Parse().
func ParseArgumentsInterspersed(f *flag.FlagSet, args []string) (params []string, err error) {
	params = []string{}
	for err == nil && len(args) > 0 {
		err = f.Parse(args)
		if err != nil {
			continue
		}
		if didStopAfterDoubleMinus(args, f.Args()) {
			params = append(params, f.Args()...)
			args = args[len(args):]
			continue
		}
		args = f.Args()
		if len(args) > 0 {
			params = append(params, args[0])
			args = args[1:]
		}
	}
	if err != nil {
		params = nil
		return
	}
	return
}

func didStopAfterDoubleMinus(args, remaining []string) bool {
	return len(args) > len(remaining) && args[len(args)-len(remaining)-1] == DoubleMinus
}
