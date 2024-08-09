/*
 * Create a version Jon's avoid routing topology in raven
 *
 * 8 gateways, connected via lan on a switch
 *
 */

topo = {
    name: "avoid_"+Math.random().toString().substr(-6),
    nodes: [
        ...["r0", "r1", "r2", "lh", "avoid", "ue1", "ue2", "dst1", "dst2"].map(x => node(x)),
    ],
    switches: [cumulus('s0')],
    links: [
        v2v("r0",   1, "s0", 1, { mac: { r0:    '04:70:00:00:02:10', s0: '04:70:00:00:01:10' } }),
        v2v("r1",   1, "s0", 2, { mac: { r1:    '04:70:00:00:02:11', s0: '04:70:00:00:01:11' } }),
        v2v("r2",   1, "s0", 3, { mac: { r2:    '04:70:00:00:02:12', s0: '04:70:00:00:01:12' } }),
        v2v("lh",   1, "s0", 4, { mac: { lh:    '04:70:00:00:02:13', s0: '04:70:00:00:01:13' } }),
        v2v("dst1", 1, "s0", 5, { mac: { dst1:  '04:70:00:00:02:14', s0: '04:70:00:00:01:14' } }),
        v2v("dst2", 1, "s0", 6, { mac: { dst2:  '04:70:00:00:02:15', s0: '04:70:00:00:01:15' } }),
        v2v("avoid",1, "s0", 7, { mac: { avoid: '04:70:00:00:02:16', s0: '04:70:00:00:01:16' } }),

        v2v("r0", 2, "ue1", 1, { mac: {   r0: '04:70:00:00:03:10', ue1: '04:70:00:00:00:10' } }),
        v2v("r1", 2, "ue1", 2, { mac: {   r1: '04:70:00:00:03:11', ue1: '04:70:00:00:00:11' } }),
        v2v("r2", 2, "ue1", 3, { mac: {   r2: '04:70:00:00:03:12', ue1: '04:70:00:00:00:12' } }),

        v2v("r0", 3, "ue2", 1, { mac: {   r0: '04:70:00:00:04:10', ue2: '04:70:00:00:05:10' } }),
        v2v("r1", 3, "ue2", 2, { mac: {   r1: '04:70:00:00:04:11', ue2: '04:70:00:00:05:11' } }),
        v2v("r2", 3, "ue2", 3, { mac: {   r2: '04:70:00:00:04:12', ue2: '04:70:00:00:05:12' } }),
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
