SERVICE=share_card_robot.service

reload: build srestart sstatus

build:
	go build -o share_card_robot .

run:
	go run .

dev:
	air

sstatus: service-status
service-status:
	sudo systemctl status $(SERVICE)

sstart: service-start
service-start:
	sudo systemctl start $(SERVICE)

sstop: service-stop
service-stop:
	sudo systemctl stop $(SERVICE)

srestart: service-restart
service-restart:
	sudo systemctl restart $(SERVICE)

sreload: service-reload
service-reload:
	sudo systemctl reload $(SERVICE)

service-enable:
	sudo systemctl enable $(SERVICE)

service-disable:
	sudo systemctl disable $(SERVICE)

service-logs:
	sudo journalctl -u $(SERVICE) -f

service-logs-tail:
	sudo journalctl -u $(SERVICE) --lines=50

service-full-status:
	sudo systemctl status $(SERVICE) --no-pager -l

hlp: help
help:
	@echo "Available commands:"
	@echo ""
	@echo "  build             - Build the binary"
	@echo "  run               - Run directly via go run"
	@echo "  dev               - Run with air (live reload)"
	@echo ""
	@echo "Service:"
	@echo "  service-start     - Start the service"
	@echo "  service-stop      - Stop the service"
	@echo "  service-restart   - Restart the service"
	@echo "  service-status    - Show service status"
	@echo "  service-logs      - Follow live logs"
