# ReadMe — MS1 Users

API REST de para la gestión de usuarios de ReadMe.

### Desarrollo local

**Variables de entorno**

```bash
cp .env.example .env
```

Edita `.env` si necesitas cambiar credenciales o puertos.

**Base de datos**

```bash
docker compose up -d
```

Levanta PostgreSQL en el puerto 5432. 

**Documentación Swagger**

Solo la primera vez (y cada vez que modifiques anotaciones en los handlers):

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init
```

> Si `swag` no se reconoce como comando, agrega el bin de Go a tu PATH y recarga la terminal:
> ```bash
> export PATH=$PATH:$(go env GOPATH)/bin
> ```

**Correr la app**

```bash
go run main.go
```

La API queda disponible en `http://localhost:8001`.
Swagger UI en `http://localhost:8001/swagger/index.html`.
