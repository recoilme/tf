# @telefeedbot

Бот в телеграм для чтения ленты вконтакте и рсс лент
Посмотреть в работе можно тут: t.me/telefeedbot

# Description
http://telegra.ph/telefeedbot-05-12


# Как это работает

Бот состоит из трех частей:
 - Сервер базы данных boltsrv
 - Сервер телеграм tgsrv
 - Сервер публикации postsrv

Для хранения опубликованных ссылок используется redis.

Собираем сервера:

```
cd boltssrv
go install

cd tgsrv 
go install

cd postsrv 
go install
```

Стартуем: ./boltsrv &

Поднимется http интерфейс на 5000 порту к базе данных (в качестве движка испоьзуется boltdb)
Возможно потребуется поднять лимит одновременно открытых соединений - ulimit


Заводим в телеграм бота.
Кладем в корень файл telefeed.bot c токеном нашего бота

Стартуем сервер телеграм: ./tgsrv>>tgsrv.log &

После этого идем в бота и активируем его - пробуем подписаться на что нибудь, например шлем боту vc.ru
Вобщем как обычно в @telefeedbot


Затем нам надо запустить сервер парсинга и рассылки

Стартуем  так: 
```
./postsrv publics>>publics.log &
./postsrv feeds>>feeds.log &
```
Не забыть поднять редис предварительно - либо заменив редис на хранение в инмемори, например, если редис претит

Для вконтакте - получите токен. см https://github.com/recoilme/tf/tree/master/vkapi
