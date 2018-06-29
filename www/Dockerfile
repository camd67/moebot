FROM nginx:1.15
COPY nginx/nginx.conf /etc/nginx/nginx.conf
COPY nginx/moebot.moe /etc/nginx/conf.d/moebot.moe
# switch this during production, maybe another dockerfile?
COPY /static/ /www/static/
#VOLUME /static /www/static
