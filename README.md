# getpages

[![codecov][cc1]][cc2]
 [![GoCard][gc1]][gc2]
 [![GitHub Release][gr1]][gr2]
 [![GitHub license][gl1]][gl2]

[cc1]: https://codecov.io/gh/LeKovr/getpages/branch/main/graph/badge.svg
[cc2]: https://codecov.io/gh/LeKovr/getpages
[gc1]: https://goreportcard.com/badge/github.com/LeKovr/getpages
[gc2]: https://goreportcard.com/report/github.com/LeKovr/getpages
[gr1]: https://img.shields.io/github/release/LeKovr/getpages.svg
[gr2]: https://github.com/LeKovr/getpages/releases
[gl1]: https://img.shields.io/github/license/LeKovr/getpages.svg
[gl2]: https://github.com/LeKovr/getpages/blob/master/LICENSE

## Описание

Тестовое задание Go
Необходимо реализовать CLI-утилиту, которая реализует асинхронную обработку входящих URL из файла, переданного в качестве аргумента данной утилите.
Формат входного файла: на каждой строке – один URL. URL может быть очень много! Но могут быть и невалидные URL.

Пример входного файла:
https://myoffice.ru
https://yandex.ru

По каждому URL получить контент и вывести в консоль его размер и время обработки. Предусмотреть обработку ошибок.

