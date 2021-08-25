import datetime
import time
import requests  

PROMETHEUS = 'http://localhost:9090/'

end_of_month = datetime.datetime.today().replace(day=1).date()

last_day = end_of_month - datetime.timedelta(days=1)
duration = '[' + str(last_day.day) + 'd]'

response = requests.get(PROMETHEUS + '/metrics',
  params={
    'query': 'sum by (job)(increase(process_cpu_seconds_total' + duration + '))',
    'time': time.mktime(end_of_month.timetuple())})
results = response.json()['data']['result']

print('{:%B %Y}:'.format(last_day))
for result in results:
  print(' {metric}: {value[1]}'.format(**result))
