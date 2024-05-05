package main

import (
	"database/sql"
	"fmt"
	"html"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type pix struct {
	pix  string
	name string
}

func (p pix) String() string {
	return fmt.Sprintf("Nome: %s | PIX : %s", p.name, p.pix)
}

func main() {
	db, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS PIX (
		id SERIAL PRIMARY KEY,
		pix VARCHAR(255),
		name VARCHAR(255)
	);`)
	if err != nil {
		panic(err)
	}

	fmt.Println("Listening...")

	http.DefaultServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<body>
		
		<h2>Doação por PIX para o RS</h2>
		<p>Esse site foi feito com o objetivo de conseguir doações para afetados pelas enchentes no Rio Grande do Sul.</p>

		<button style="font-size=16px;" onclick="window.location.href='./doar'">Doar PIX</button>

		<button style="font-size=16px;" onclick="window.location.href='./pedir'">Pedir doação PIX</button>

		<p>Criador: <a href="https://www.linkedin.com/in/gabriel-seibel">Gabriel de Souza Seibel</a><p>
		<p>Pela urgência o site não tenta ser muito bonito, mas sim atender a população o mais rápido possível.<p>
		
		</body>
		</html>
		`)
	})

	http.DefaultServeMux.HandleFunc("/pedir", func(w http.ResponseWriter, r *http.Request) {
		slog.Info(html.EscapeString(r.URL.Path))

		fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<body>
		
		<h2>Peça um Doação por PIX</h2>
		<p>Você está precisando de doação? Coloque seus dados abaixo e clique em pedir!</p>
		<p>Os pedidos podem ser atendidos por voluntários, apesar de que não há garantias.</p>
		
		<form action="/pedido">
		  <label for="name">Nome:</label><br>
		  <input type="text" id="name" name="name" hint="Seu nome aqui"><br>
		  <label for="pix">PIX:</label><br>
		  <input type="text" id="pix" name="pix" hint="Sua chave PIX aqui"><br><br>
		  <input type="submit" value="Pedir" style="font-size=16px;">
		</form> 
		
		</body>
		</html>
		`)
	})

	http.DefaultServeMux.HandleFunc("/pedido", func(w http.ResponseWriter, r *http.Request) {
		p := pix{
			r.FormValue("name"),
			r.FormValue("pix"),
		}
		slog.Info(p.String())

		_, err := db.Exec(`INSERT INTO PIX (name, pix) VALUES ($1, $2);`, p.name, p.pix)
		if err != nil {
			slog.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<body>
		
		<p>Pedido registrado com sucesso.</p>
		
		<button style="font-size=16px;" onclick="window.location.href='../'">Voltar a pagina inicial</button>

		</body>
		</html>
		`)
	})

	http.DefaultServeMux.HandleFunc("/doar", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(html.EscapeString(r.URL.Path))

		rows, err := db.Query(`SELECT pix, name FROM PIX ORDER BY RANDOM() LIMIT 100;`)
		if err != nil {

			if err == sql.ErrNoRows {
				slog.Error(err.Error())
				fmt.Fprintf(w, "Sem dados")
				w.WriteHeader(http.StatusOK)
				return
			}

			slog.Error(err.Error())
			fmt.Fprintf(w, "Erro no banco de dados")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		list := make([]string, 0)
		i := 0
		defer rows.Close()
		for rows.Next() {
			var result pix
			err := rows.Scan(&result.pix, &result.name)
			if err != nil {
				slog.Error(err.Error())
				return
			}

			list = append(list, fmt.Sprintf("<li>%s</ls>", result.String()))
			i++
		}
		slog.Info(strconv.Itoa(i))

		fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<body>
		
		<h2>Faça um PIX</h2>
		<p>Tendo em vista as enchentes no Rio Grande do Sul, esta página lista pessoas precisando de doação via PIX.</p>
		<p>Copie o PIX de uma pessoa listada abaixo e cole no seu app bancário para doar.</p>
		
		<ul>%s</ul>
		
		</body>
		</html>		
		`, strings.Join(list, ""))
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
