#!/bin/bash
source test/define.sh $0 $1

rm -rf *1111*
VPortA="-v3 --port 1111"
VPortB="-v3 --port 1105"

RunTest create_wallet "${VPortA} --specified 0" 's#.*: \(.*\)#\1#g'
AddressA="${TestRegMatch}"

RunTest create_wallet "${VPortB} --specified 0" 's#.*: \(.*\)#\1#g'
AddressB="${TestRegMatch}"

RunTest create_blockchain "${VPortA} --address ${AddressA}"

RunTest all "${VPortA} --address ${AddressA}" &
sleep 1

RunTest sync "${VPortB} --address ${AddressB}"

killall_blockchain() {
	sleep 3
	killall blockchain
}
killall_blockchain &

RunTest mining "${VPortB} --address ${AddressB}"

RunTest get_version "${VPortA} --address ${AddressA}"
