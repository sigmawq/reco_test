# Asana periodic extractor

Basically a program to periodically query things from Asana and store it locally.

## Installation

1. Have go compiler installed on the target system
2. cd to source directory
3. ```go build .```
4. Run the target executable

## Config docs
| Param   | Description                                                        |
|---------|--------------------------------------------------------------------|
| asana_secret | Your asana API key                                                 |
| extraction_period_seconds | Extractor period                                                   |
| retry_count   | When calling asana API, this will be the maximum amount of retries |
| retry_period_milliseconds   | When calling asana API, this will be the delay between calls       |
| asana_max_rpm   | Maximum asana RPM before slowing down                              |

## License

[MIT](https://choosealicense.com/licenses/mit/)