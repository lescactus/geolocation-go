#!/usr/bin/python

from locust import HttpUser, task, between
from faker import Faker
import random

def random_fixed_ip():
    ip = [
        "178.167.175.30",
        "114.24.132.31",
        "91.205.163.212",
        "179.6.254.234",
        "216.196.189.137",
        "217.44.226.187",
        "155.211.125.64",
        "191.194.46.98",
        "63.127.26.70",
        "154.100.243.131"
    ]

    return random.choice(ip)

def random_ip():
    faker = Faker()
    return faker.ipv4()


class QuickstartUser(HttpUser):
    wait_time = between(1, 2)

    @task(4)
    def fixed_ip(self):
        self.client.get("/rest/v1/1.2.3.4")
        self.client.get("/rest/v1/4.5.6.7")
        self.client.get("/rest/v1/7.8.9.0")
    
    @task(3)
    def random_fixed_ips(self):
        ip = random_fixed_ip()
        self.client.get(f"/rest/v1/{ip}")

    @task
    def random_ip(self):
        ip = random_ip()
        self.client.get(f"/rest/v1/{ip}")
