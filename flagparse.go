package cli

import "flag"

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
