// Package application contém os casos de uso, que orquestram o domínio.
//
// Os ports ficam em application/ports: interfaces de entrada (casos de uso que os
// adapters de entrada chamam) e de saída (repositórios, event store, cache), que os
// adapters de saída implementam. Pode importar domain; nunca importa adapters.
package application
