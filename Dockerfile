#FROM ubuntu:latest
#LABEL authors="xuver"
#
#ENTRYPOINT ["top", "-b"]


FROM golang:1.20-alpine

WORKDIR /build

#Копирование go.mod и go.sum
COPY go.mod go.sum ./

#Загрузка независимости
RUN go mod download


# Копируем все файлы проекта в контейнер ???
COPY . .

# Компилируем Go-приложение
RUN go build -o main .

# Указание порта, который будет слушать приложение
EXPOSE 8080

# Команда для запуска приложения
CMD ["./main"]