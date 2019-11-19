#define wait(s) atomic { s > 0 -> s-- }
#define signal(s) s++

mtype = { NEW_HOST, ADD_HOST, NEW_BLOCK, SET_BLOCKS, ADD_BLOCK }
mtype handle = NEW_HOST;

byte S = 1
byte cc = 0

/*
    1) Esta función representa al handle que atiende todas nuestras peticiones en el nodo
    2) Cada vez que llega una petición a nuestro nodo, se genera un "GO handle" por lo que son ejecuciones paralelas trabajando con la misma
    varibales globales: El arreglo de IPs y nuestra Blockchain
    3) En este archivo de promela, verificamos que nuestras actividades en la blockchain son concurrentes y que la sección critica solo
    es accedida por una solicitud a la vez. Asegurando la integridad de ambas variables globales antes mencionadas
    4) Utilizamos la función Assert para validar lo anterior, además de los comandos en Sprin.

*/
proctype H() {
	byte repeat = 0
	do
	:: repeat < 50 ->
		repeat++
        wait(S)
        cc ++
		if
		:: handle == NEW_HOST -> 	handle = ADD_HOST
		:: handle == ADD_HOST -> 	handle = NEW_BLOCK
		:: handle == NEW_BLOCK -> 	handle = SET_BLOCKS
        :: handle == SET_BLOCKS -> 	handle = ADD_BLOCK
        :: handle == ADD_BLOCK -> 	handle = NEW_HOST
		fi
        assert(cc < 2)
        cc--
		printf("The handle is now work with task %e\n", handle)
        signal(S)
	:: else -> break
	od
}


init {
	atomic {
		run H()
		run H()
        run H()
	}
	(_nr_pr == 1) -> printf("cc = %d\n", cc)
}