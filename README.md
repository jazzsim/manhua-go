A go backend for scraping http server page and sort them into folders and files. It supports basic auth for http server.

Exmaple request body:
```
{
  "url": "http://192.168.0.1:5173/",
  "username": "admin",
  "password": "C0mplic@t3dP@ssW0rd"
}
```


Example response:
```
{
    "folders": [
        "../",
        "books/",
        "cooking/",
        "health/",
        "movies/",
        "music/"
    ],
    "files": [
        {
            "media": "Audio",
            "filename": "Love_Yourself.mp3"
        },
        {
            "media": "Video",
            "filename": "Planet_Lockdown.mp4"
        },
        {
            "media": "Video",
            "filename": "Uninformed_Consent.mp4"
        },
        {
            "media": "Others",
            "filename": "forHarold.html"
        },
        {
            "media": "Others",
            "filename": "index-original.html"
        },
        {
            "media": "Others",
            "filename": "x10.php"
        },
        {
            "media": "Others",
            "filename": "x10.php3"
        }
    ]
}
```
