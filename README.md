# Web crawler

![](/preview.png)


### Запуск

Для запуску додатка в docker-контейнерах, можете запустити файл `build-services.bat` або `build-services.sh` в корені проекту для компілювання бінарних файлів, та виконати команду: `docker-compose up --build`.
Або можете завантажити вже скомпільовані файли з [цієї папки](https://www.dropbox.com/scl/fo/9uxrpkjusfsgto4icgrwu/ADUbNKBAB44OO5yN1J79ntM?rlkey=9jz1pdsm1sjgu219idgk25gm1&st=oajvueq0&dl=0) та запустити команду в ній.

*Проект орієнтовано буде остаточно запущений і готовий до роботи протягом 2-3 хвилин після завантаження та побудови всіх образів.*

Після запуску додатка будуть доступні такі URL:
- http://localhost:8000 - переходячи за цим посиланням відобразиться веб-інтерфейс для взаємодії з додатком (Guest service).
- http://localhost:8001 - надає API для керування сервісами (Controller service).

Для швидкого тесту додатка з не великою кількістю посилань, рекомендую наступні сайти с використанням флагу "Same domain only":
- https://send-anywhere.com
- https://translate.google.com
- https://www.nohello.com
- https://www.tldrthis.com

---


# Опис

Сервіс призначений для рекурсивного збору даних (ключових слів) із веб-сторінок. Починаючи із заданого юзером URL, витягує із сторінки дані та посилання, потім переходить за цими посиланнями, повторюючи процес, доки всі пов'язані сторінки не будуть вивчені, включаючи SPA додатки. Дослідження сторінок також можна обмежити, фільтруючи посилання по імені домену стартового URL.
Додаток створений з використанням мікросервісної архітектури. Кожен мікросервіс іденпотентний, відмовостійкий та масштабований.

### Проект складається з таких мікросервісів:
- **Guest (Main)** - надає API для створення, читання і видалення запитів, та надає веб-інтерфейс для зручного користування.
- **Worker** - обробляє посилання, тобто здійснює парсинг ключових слів і інших посилань, на основі яких створює нові задачі для воркерів, та оповіщуює агрегатор, якщо задачі потенційно закінчилися.
- **Aggregator** - перевіряє чи закінчилися задачі для воркерів для вказаного запиту, і якщо так, то агрегує результати в базу даних.
- **Controller** - надає API для управління іншими сервісами та здійснює їх автоматичне масштабування.

Сторонні сервіси:
- **PostgreSQL**
- **Redis**
- **Kafka**
- **ElasticSearch**
- **Logstash**
- **Kibana**

---


# Основні фічі

- Рекурсивний пошук даних на сайті;
- Присутній веб-інтерфейс для взаємодії з додатком;
- Парсинг SPA-додатків за рахунок використання браузера Chromium;
- Обробка відносних посилань;
- Автоматичне масштабування мікросервісів в залежності від напливу трафіку/задач;
- Планове масштабування мікросервісів;
- Централізоване управління мікросервісами через API;
- Мікросервісів оснащені логуванням з підтримкою ELK-стеку;
- Всі задачі мікросервісів іденпотентні;
- Експонінційне збільшення паузи ретрая в разі збоїв зв'язку між мікросервісами;
- Є можливість запустити мікросервіси в багатопоточному режимі;
- Висока відмовостійкість завдяки гнучким механізмам обробки збоїв;
- Висока швидкодія додатку завдяки мові програмування Golang;
- Легкий запуск і налаштування мікросервісів в docker-контейнерах;

---


# Технології і залежності

### Back-end:
- Golang 1.23.1
- Echo 4.12.0 (github.com/labstack/echo/v4)
- GORM 1.25.12 (gorm.io/gorm)
- GORM PosgreSQL driver 1.5.9 (gorm.io/driver/postgres)
- Kafka client 0.4.47 (github.com/segmentio/kafka-go)
- Redis client 9.7.0 (github.com/redis/go-redis/v9)
- ROD 0.116.2 (github.com/go-rod/rod)

### Services
- PostgreSQL 17.0
- Redis 5.0.14
- Kafka 3.8.0 (Zookeeper 3.8.0)
- ElasticSearch 8.10.2
- Logstash 8.10.2
- Kibana 8.10.2

---


# Структура файлів

- **main.go** - головний файл, з якого здійснюється запуск мікросервісу.
- **tools/** - зберігає універсальні допоміжні інструменти, які використовуються по всьому мікросервісу.
- **cache/** - містить файли та інтерфейси для роботи з кешем.
- **msg_broker/** - містить файли та інтерфейси для роботи брокером повідомлень.
- **logger/** - містить файли та інтерфейси для роботи логами.
- **database/** - містить файли, моделі та інтерфейси для роботи з базою даних.
- **static/** - містить статичні файли мікросервісу.
- **docker/** - містить файли для конструювання docker-образу.
- **app/ worker/ aggregator/** - містить основну структуру (ядро) мікросервісу.
- **runner.bat(.sh), builder.bat(.sh)** - файли для автоматичного запуску та компілингу (під Linux) мікросервісу.

---


# Мікросервіси:

### Guest (Main) service
- **/ (GET)** - повертає html файл для взаємодії з додатком..
- **/request (POST)** - створює запит.
- **/request (GET)** - повертає результати запиту по вказаному в параметрах ідентифікатору.
- **/request (DELETE)** - видаляє запит по вказаному в параметрах ідентифікатору.

Більш детально дізнатися про API можна в файлі [api_docs.md](api_docs.md).

**Змінні оточення:**
- `ID` - ідентифікатор екземпляру мікросервіса. За замовченням: 1.
- `POSTGRES_CONN` - рядок для налаштування СУБД PostgreSQL в форматі *key=value* через пробіл. Вказується хост (`host`), порт (`post`), ім'я користувача (`user`) та пароль (`password`) для входу, назва бази даних (`dbname`) та чи потрібно використовувати шифрування (`sslmode`).
- `REDIS_HOST` - рядок для налаштування Redis в форматі *host:port*.
- `KAFKA_HOSTS` - рядок для налаштування Kafka в форматі *host1:port1,host2:port2,...*.
- `LOGS_DIR` - папка в форматі *path/to/logs/*, в якій будуть зберігатися файли логів мікросервісу. Якщо дана змінна не буде задана, то логи не будуть зберігатися в файлі.
- `LOGSTASH_HOST` - URL, по якому мікросервісу буде відправляти логи в Logstash. Якщо дана змінна не буде задана, то логи не будуть відправлятися в Logstash.


### Worker service

**1. Отримання задачі:** Воркер отримує і валідує задачу від Kafka з топіку `FindWords`, в якій в форматі JSON передається наступне:
- Посилання (`url`).
- Ідентифікатор запиту, до якого відноситься дане посилання (`request_id`).
- Кількість невдалих спроб обробки цього посилання (`attempts`).

**2. Перевірка посилання:**
- Перевіряється в Redis на наявність даного посилання в списку вже оброблених посилань `completed_urls`.
- Якщо посилання вже було оброблено, то створюється задача для Aggregator service і відправляється в топік `AggregateResult`, після чого обробка завершується.

Це зроблено для того, щоб запобігти втрату даних під час збою роботи воркера, так як раніше він міг відмітити посилання як оброблене, але не встигнути викликати агрегатор для перевірки кінцевого результату запиту і запит ніколи не буде оброблений.

**3. Завантаження документа по посиланню:**
- За допомогою бібліотеки ROD та браузера Chromium здійснюється завантаження сторінки за посиланням.
- Якщо перехід за посиланням невдалий, а кількість невдалих спроб даної задачі не перевищує 2 рази, то до поля `attempts` додається ще одна невдала спроба, і задача знову додається в чергу задач `FindWords`.
- Якщо перехід невдалий і кількість невдалих спроб дорівнює або більше 2-х разів, то URL додається в список `completed_urls`, але зі статусом `Fail`, після чого викликається агрегатор і здійснюється завершення обробки.
- Якщо перехід вдалий, але сторінка не змогла завантажилась (коли це SPA додаток), то посилання також додається в список оброблених посилань, але зі статусом `Unreaded` і воркер також оповіщує агрегатор та завершує обробку.

**4. Обробка документа:**
- Якщо сторінка завантажилась коректно, то потім із Redis дістаються дані для обробки запиту із списку `requests_data`, тобто ключові слова, які потрібно знайти, і чи потрібно фільтрувати посилання за доменом.
- Потім проходить парсинг посилань та створення нових задач:
    - Здійснюється пошук всіх посилань в документі.
    - Всі знайдені посилання фільтруються (перевіряючи чи є посилання в списку `all_urls`), залишаючи тільки ті, що ще не були знайдені раніше.
    - На основі відфільтрованих посилань створюються задачі та відправляються в `FindWords`.
    - Посилання додаються в список всіх знайдених посилань `all_urls` в Redis.
- За тим здійснюється парсинг ключових слів, і обробляєме посилання (як ключ) разом із цими словами та статусом `Success` (як значення), додається до словаря оброблених посилань `completed_urls`.

**5. Оповіщення агрегатора:** Якщо кількість відфільтрованих задач дорівнює нулю, то дане обробляєме посилання вірогідно є потенційно останнім в цьому запиті, тому створюється і відсилається задача для агрегатора в `AggregateResult` для перевірки цієї гіпотези, і в разі її підтвердження здійснюється агрегації даних.

**Змінні оточення:**
- `ID` - ідентифікатор екземпляру мікросервіса. За замовченням: 1.
- `SEARCH_CHROMIUM` - якщо вказано як `yes`, то сервіс не завантажує, а шукає браузер в системі.
- `TASK_RECEPIENTS` - кількість потоків, які обробляють задачі. За замовченням один.
- `REDIS_HOST` - рядок для налаштування Redis в форматі *host:port*.
- `KAFKA_HOSTS` - рядок для налаштування Kafka в форматі *host1:port1,host2:port2,...*.
- `LOGS_DIR` - папка в форматі *path/to/logs/*, в якій будуть зберігатися файли логів мікросервісу. Якщо дана змінна не буде задана, то логи не будуть зберігатися в файлі.
- `LOGSTASH_HOST` - URL, по якому мікросервісу буде відправляти логи в Logstash. Якщо дана змінна не буде задана, то логи не будуть відправлятися в Logstash.


### Aggregator service

**1. Отримання задачі:** Сервіс отримує задачу від Kafka з топіку `AggregateResult`, у якій передається ідентифікатор запиту, який потрібно перевірити.

**2. Перевірка виконання всіх задач:** Сервіс перевіряє, чи всі посилання, пов'язані з запитом, були оброблені. Для цього з Redis отримуються дані:
- Кількість всіх знайдених посилань (`all_urls`).
- Кількість вже оброблених посилань (`completed_urls`).

Якщо кількість знайдених і оброблених посилань збігається, тобто задачі для цього запиту закінчилися, то запит вважається завершеним, інакше задача ігнорується.

**3. Агрегація даних:** Якщо запит дійсно був завершений, то здійснюється агрегація даних:
- Сервіс дістає результати обробки з `completed_urls`.
- Приводить їх до потрібного формату.
- Зберігає агреговані результати у базу даних і відзначає запит в БД як оброблений.

**4. Очищення даних:**
Після успішної агрегації сервіс видаляє всі пов'язані з обробленим запитом дані із Redis.

**Змінні оточення:**
- `ID` - ідентифікатор екземпляру мікросервіса. За замовченням: 1.
- `TASK_RECEPIENTS` - кількість потоків, які обробляють задачі. За замовченням: 1.
- `POSTGRES_CONN` - рядок для налаштування СУБД PostgreSQL в форматі *key=value* через пробіл. Вказується хост (`host`), порт (`post`), ім'я користувача (`user`) та пароль (`password`) для входу, назва бази даних (`dbname`) та чи потрібно використовувати шифрування (`sslmode`).
- `REDIS_HOST` - рядок для налаштування Redis в форматі *host:port*.
- `KAFKA_HOSTS` - рядок для налаштування Kafka в форматі *host1:port1,host2:port2,...*.
- `LOGS_DIR` - папка в форматі *path/to/logs/*, в якій будуть зберігатися файли логів мікросервісу. Якщо дана змінна не буде задана, то логи не будуть зберігатися в файлі.
- `LOGSTASH_HOST` - URL, по якому мікросервісу буде відправляти логи в Logstash. Якщо дана змінна не буде задана, то логи не будуть відправлятися в Logstash.


### Controller service
- **/sign (POST)** - посилає вказаний сигнал сервісам.

Більш детально дізнатися про API можна в файлі [api_docs.md](api_docs.md).

**Змінні оточення:**
- `PASSWORD_HASH` - захешований за алгоритмом SHA256 в нижньому регістрі пароль для доступу через API.
- `ENABLE_PLANNER` - якщо вказано як `yes`, то сервіс здійснює автоматичне масштабування не аналізуючи трафік/задачі, а по заданому в змінній `PLANNER_RULES` графіку.
- `PLANNER_RULES` - правила планувальника що до масштабування мікросервісів в форматі *start_timemark1,end_timemark1,service1:number_of_service1,service2:number_of_service2,...@start_timemark2,end_timemark2,service1:number_of_service1,service2:number_of_service2*. Формат timemark:
    - `M` - місяць;
    - `D` - дні;
    - `W` - дні тижня;
    - `H` - години.
Приклад: `W6.H16,W7.H22,worker:10,main:3@W1.H8,W5.H12,worker:5,aggregator:1`, тут вказано два правила:
    - По вихідним (`W6-W7`) з 16:00 до 22:00 (`H16-H22`) працює 10 воркерів і 3 гостевих сервісів;
    - По будням (`W1-W5`) з 08:00 до 12:00 (`H8-H12`) працює 5 воркерів і 1 агрегатор;
- `KAFKA_HOSTS` - рядок для налаштування Kafka в форматі *host1:port1,host2:port2,...*.
- `LOGS_DIR` - папка в форматі *path/to/logs/*, в якій будуть зберігатися файли логів мікросервісу. Якщо дана змінна не буде задана, то логи не будуть зберігатися в файлі.
- `LOGSTASH_HOST` - URL, по якому мікросервісу буде відправляти логи в Logstash. Якщо дана змінна не буде задана, то логи не будуть відправлятися в Logstash.