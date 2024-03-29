# Spoti2VK

Скрипт транслирующий музыку играющую у вас в Spotify в ваш статус ВКонтакте в формате проигрываемой аудиозаписи.

![](https://lolipa.in/static/img/opera_jqfdacglKV.png)
 
Для работы скрипта необходимо установить пакет **GoFrequency**

Установка, со стандартным GOROOT, GOPATH итд. 

`go get github.com/p2love/GoFrequency/GoFrequency`

<blockquote>(а если вы их меняли, то знали на что идете и сами сможете поставить этот пакет)</blockquote>

Для получения текущего статуса в Spotify, в ваших настройках приватности аккаунта должна быть включена данная возможность.

Сейчас период обновления статуса составляет 10 секунд, что является тестовым вариантом и не принимается в качестве баг-репорта.

Также корректность передачи NP в статус не принимается без конкретизации и шагов воспроизведения.

### Подготовка к работе

Для работы необходимо получить AuthCode от Spotify со Scope = user-read-currently-playing и **токен VK с доступом к API audio методам.**

Заполняются в `const SpotiAuthCode` и `const VKToken` соотвественно.

Более подробная инструкция находится в комментариях внутри файла `Spoti2VK.go`


### Запуск и работа

Находясь в директории с файлом скрипта:
```
go mod init Spoti
go mod tidy

go run Spoti2VK.go
```


## О найденных багах сообщать в раздел Issue:
https://github.com/P2LOVE/Spoti2VK/issues

Для более подробных баг-репортов, желательно включить DebugMode. Так как на GoLang не получается хэндлить паники, остается только вариант логгирования в консоль всех треков, перед попыткой их включить.

Если вы поймаете panic, то в nohup.out, перед ним, будет последний трек, который, скорее всего, и вызвал этот паник.
