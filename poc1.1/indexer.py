from elasticsearch.helpers import bulk

import devices
import math

def index(es, num, iname, schema):
    bulk(es, devices.gen(num, iname, schema))
