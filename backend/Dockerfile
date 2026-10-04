FROM maven:3.9-eclipse-temurin-21 AS build
WORKDIR /app

COPY backend/pom.xml .
RUN mvn -B -ntp dependency:go-offline

COPY backend/src ./src
RUN mvn -B -ntp package -DskipTests

FROM eclipse-temurin:21-jre-jammy
WORKDIR /app

RUN groupadd -r pulse && useradd -r -g pulse pulse

COPY --from=build /app/target/pulse-0.0.1-SNAPSHOT.jar app.jar
RUN chown -R pulse:pulse /app
USER pulse

EXPOSE 8080

ENTRYPOINT ["java", \
  "-XX:MaxRAMPercentage=65", \
  "-XX:+UseSerialGC", \
  "-XX:MaxMetaspaceSize=96m", \
  "-XX:ReservedCodeCacheSize=32m", \
  "-XX:MaxDirectMemorySize=32m", \
  "-Xss512k", \
  "-Djava.security.egd=file:/dev/./urandom", \
  "-jar", "app.jar"]