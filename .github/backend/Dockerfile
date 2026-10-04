FROM maven:3.9-eclipse-temurin-21 AS build
WORKDIR /app

COPY pom.xml .
RUN mvn -B -ntp dependency:go-offline

COPY src ./src
RUN mvn -B -ntp package -DskipTests

FROM eclipse-temurin:21-jre-jammy
WORKDIR /app

RUN groupadd -r pulse && useradd -r -g pulse pulse

COPY --from=build /app/target/pulse-0.0.1-SNAPSHOT.jar app.jar
RUN chown -R pulse:pulse /app
USER pulse

EXPOSE 8080

ENTRYPOINT ["java", \
"-XX:MaxRAMPercentage=70", \
"-XX:+UseSerialGC", \
"-Djava.security.egd=file:/dev/./urandom", \
"-jar", "app.jar"]