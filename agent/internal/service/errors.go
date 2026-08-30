// Package service coordena os collectors e prepara os models para a API.
package service

import "errors"

// ErrUnavailable indica que uma métrica não pôde ser obtida agora, mas pode
// voltar a funcionar. A camada HTTP a traduz em 503, sinalizando ao cliente
// que vale tentar de novo — diferente de um erro definitivo.
var ErrUnavailable = errors.New("métrica temporariamente indisponível")

// ErrInvalidArgument indica pedido malformado — um parâmetro fora do domínio
// aceito. A camada HTTP a traduz em 400: repetir a mesma requisição não vai
// funcionar, quem precisa mudar é o cliente.
var ErrInvalidArgument = errors.New("parâmetro inválido")
