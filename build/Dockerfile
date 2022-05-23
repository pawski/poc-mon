FROM alpine:3.15.4

ADD build/bin/pocmon_linux_64 /var/lib/pocmon
ADD ../app.yml.dist /var/lib/app.yml
ADD ../env.yml.dist /var/lib/env.yml

EXPOSE 8000
ENTRYPOINT /var/lib/pocmon