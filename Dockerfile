FROM golang:1.3.1-onbuild
EXPOSE 3000
CMD app --clientid $CLIENTID --secret $SECRET --calendar $CALENDAR

#docker build -t my-golang-app --rm .
#docker run -it --rm --name my-running-app my-golang-app
