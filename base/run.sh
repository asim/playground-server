#!/bin/bash

# Run a program via the following options
# 
# A file name passed in on command line
# A Procfile found in /code
# A main.{ext} file found in /code

DIR=/code

run() {
	if [ ! -f $1 ]; then
		echo "program not found"
		exit 1
	fi

	program=$1
	extension="${program##*.}"

	case "$extension" in
		"sh")
			bash $program
			;;
		"c")
			gcc $program && ./a.out
			;;
		"go")
			go run $program
			;;
		"pl")
			perl $program
			;;
		"py")
    			python $program
			;;
		"rb")
			ruby -e STDOUT.sync=true -e 'load($0=ARGV.shift)' $program
			;;
		*)
			echo "invalid language"
			exit 1
		;;
	esac
}

if [ -n "$1" ]; then
	# run passed in filename
	run $1
elif [ -f Procfile ]; then
	# run procfile
	echo "Found Procfile"
	command=`awk -F : '{print substr($0, index($0, $2))}' Procfile`
	exec bash -c $command
else
	# try find a main file
	echo "Searching for main program"
	for file in $(ls $DIR/main.* 2>/dev/null); do
		run $file
	done
fi


