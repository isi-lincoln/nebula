/*
 * Create a version Jon's avoid routing topology in raven
 *
 * 8 gateways, connected via lan on a switch
 *
 */

topo = {
    name: "avoid_"+Math.random().toString().substr(-6),
    nodes: [
        ...["r0", "r1", "lh", "ue", "dst"].map(x => node(x)),
    ],
    switches: [cumulus('s0')],
    links: [
        v2v("r0", 1, "s0", 1, { mac: {   r0: '04:70:00:00:02:10', s0: '04:70:00:00:01:10' } }),
        v2v("r1", 1, "s0", 2, { mac: {   r1: '04:70:00:00:02:11', s0: '04:70:00:00:01:11' } }),
        v2v("lh", 1, "s0", 3, { mac: {   lh: '04:70:00:00:02:12', s0: '04:70:00:00:01:12' } }),
        v2v("dst",1, "s0", 4, { mac: {   lh: '04:70:00:00:02:13', s0: '04:70:00:00:01:13' } }),
        v2v("r0", 2, "ue", 1, { mac: {   r0: '04:70:00:00:03:10', ue: '04:70:00:00:00:10' } }),
        v2v("r1", 2, "ue", 2, { mac: {   r1: '04:70:00:00:03:11', ue: '04:70:00:00:00:11' } }),
    ]
}

function node(name) {
    return {
        name: name,
	defaultnic: 'e1000',
        image: 'ubuntu-2204',
        mounts: [{ source: env.PWD+"/../..", point: "/avoid" }], 
        cpu: { cores: 2 },
        memory: { capacity: GB(4) },
    };
}

function cumulus(name) {
    return {
        name: name,
        image: 'cumulusvx-4.1',
        cpu: { cores: 1 },
        memory: { capacity: GB(1) },
     }
}

function v2v(a, ai, b, bi, props={}) {
    lnk = Link(a, ai, b, bi, props);
    lnk.v2v = true;
    return lnk;
}
