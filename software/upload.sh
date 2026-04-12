set -e
arduino-cli compile /home/luccas/pon-bot/software/driver -b teensy:avr:teensy41
arduino-cli upload /home/luccas/pon-bot/software/driver -p /dev/ttyACM0 -b teensy:avr:teensy41
pushd supervisor
go build

sleep 1
echo "Starting supervisor."
sudo ./supervisor
popd