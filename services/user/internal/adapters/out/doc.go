// Package out agrupa os adapters de saída: implementações dos ports de saída
// definidos em application/ports (ex.: out/postgres, out/redis, out/kurrentdb).
//
// Cada adapter fica num subpacote próprio e converte entre os tipos do domínio e os
// tipos da tecnologia (linhas SQL, eventos serializados). Um adapter não chama outro.
package out
