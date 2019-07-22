#/bin/bash

if [[ `uname` == 'Darwin' ]]; then
  #to use gnome command line tools you have to install Homebrew
  #in mac os the gnome command ’readlink’ is ’greadlink’
  export PATH="/usr/local/bin:$PATH"
  Script=$(greadlink -f $0)
  Scriptpath=`dirname $Script`
  Root=$(greadlink -f $Scriptpath/../../)
else
  Script=$(readlink -f $0)
  Scriptpath=`dirname $Script`
  Root=$(readlink -f $Scriptpath/../../)
fi

function failed {
  echo "$1" 1>&2
  exit 1;
}

Branch="master"
if [ -n "$1" ]; then
  Branch=$1
fi
Revision=$2

cd $Root

echo "Start update to origin/$Branch"

# Reset local changes
git reset HEAD --hard || failed "Error while reseting repository"

# Remove all untracked files and directories
git clean -f -d || failed "Error while deleting untracked files"

# Get latest remote
git fetch origin $Branch || failed "Error while fetching branch $Branch" 

# Switch to branch
git checkout $Branch || failed "Error while checkout branch $Branch"

# Pull, now that we are on the correct branch
git pull || failed "Error while pulling" 

if [ -n "$Revision" ]; then
	# Get specific revition
	git checkout $Revision . || failed "Error qhile checkout revision"
	echo "Checkout revision: " $Revision
fi

git lfs pull || failed "Error while lfs pull"