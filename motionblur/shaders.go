package main

// TODO: Clean these up.

const vertexSource = `//#version 120 // OpenGL 2.1.
//#version 100 // WebGL.

attribute vec3 aVertexPosition;

uniform mat4 uMVMatrix;
uniform mat4 uPMatrix;

varying vec4 verpos;

void main() {
	gl_Position = uPMatrix * uMVMatrix * vec4(aVertexPosition, 1.0);
	verpos = vec4(aVertexPosition, 1.0) * uMVMatrix;
}
`

const fragmentSource = `//#version 120 // OpenGL 2.1.
//#version 100 // WebGL.

#ifdef GL_ES
	precision highp float;
#endif

uniform vec3 tri0v0;
uniform vec3 tri0v1;
uniform vec3 tri0v2;
uniform vec3 tri1v0;
uniform vec3 tri1v1;
uniform vec3 tri1v2;

varying vec4 verpos;

struct Line3
{
	vec3 Origin;
	vec3 Direction;
};

struct Triangle3
{
	vec3 V[3];
};

bool IntrLine3Triangle3_Find(Line3 line, Triangle3 triangle, out float TriBary[3])
{
	// Compute the offset origin, edges, and normal.
	vec3 diff = line.Origin - triangle.V[0];
	vec3 edge1 = triangle.V[1] - triangle.V[0];
	vec3 edge2 = triangle.V[2] - triangle.V[0];
	vec3 normal = cross(edge1, edge2);

	// Solve Q + t*D = b1*E1 + b2*E2 (Q = diff, D = line direction,
	// E1 = edge1, E2 = edge2, N = Cross(E1,E2)) by
	//   |Dot(D,N)|*b1 = sign(Dot(D,N))*Dot(D,Cross(Q,E2))
	//   |Dot(D,N)|*b2 = sign(Dot(D,N))*Dot(D,Cross(E1,Q))
	//   |Dot(D,N)|*t = -sign(Dot(D,N))*Dot(Q,N)
	float DdN = dot(line.Direction, normal);
	float sign;
	if (DdN > 0.0)	///Math<float>::ZERO_TOLERANCE
	{
		sign = 1.0;
	}
	else if (DdN < -0.0)	///Math<float>::ZERO_TOLERANCE
	{
		sign = -1.0;
		DdN = -DdN;
	}
	else
	{
		// Line and triangle are parallel, call it a "no intersection"
		// even if the line does intersect.
		///mIntersectionType = IT_EMPTY;
		return false;
	}

	float DdQxE2 = sign * dot(line.Direction, cross(diff, edge2));
	if (DdQxE2 >= 0.0)
	{
		float DdE1xQ = sign * dot(line.Direction, cross(edge1, diff));
		if (DdE1xQ >= 0.0)
		{
			if (DdQxE2 + DdE1xQ <= DdN + 0.03)	// HACK: Low precision fix.
			{
				// Line intersects triangle.
				///float QdN = -sign * dot(diff, normal);
				float inv = 1.0 / DdN;
				///lineParameter = QdN * inv;
				TriBary[1] = DdQxE2*inv;
				TriBary[2] = DdE1xQ*inv;
				TriBary[0] = 1.0 - TriBary[1] - TriBary[2];
				///mIntersectionType = IT_POINT;
				return true;
			}
			// else: b1+b2 > 1, no intersection
		}
		// else: b2 < 0, no intersection
	}
	// else: b1 < 0, no intersection

	return false;
}

bool IntrLine3Triangle3_Find(Line3 line, Triangle3 triangle, float tmax, vec3 velocity0, vec3 velocity1, out float ContactTime)
{
	float TriBary[3];

	if (IntrLine3Triangle3_Find(line, triangle, TriBary))
	{
		ContactTime = 0.0;
		return true;
	}
	else
	{
		// Velocity relative to line
		vec3 relVelocity = (velocity1 - velocity0) * tmax;

		Triangle3 triangle1;
		triangle1.V[0] = triangle.V[0] + relVelocity;
		triangle1.V[1] = triangle.V[1] + relVelocity;
		triangle1.V[2] = triangle.V[2] + relVelocity;

		float ClosestContactTime = 2.0;
		{
			float TriBary[3];

			{
				Triangle3 tri;
				tri.V[0] = triangle.V[0];
				tri.V[1] = triangle1.V[0];
				tri.V[2] = triangle1.V[1];

				if (IntrLine3Triangle3_Find(line, tri, TriBary)) {
					ClosestContactTime = min(ClosestContactTime, TriBary[1] + TriBary[2]);
				}
			}

			{
				Triangle3 tri;
				tri.V[0] = triangle.V[0];
				tri.V[1] = triangle.V[1];
				tri.V[2] = triangle1.V[1];

				if (IntrLine3Triangle3_Find(line, tri, TriBary)) {
					ClosestContactTime = min(ClosestContactTime, TriBary[2]);
				}
			}

			{
				Triangle3 tri;
				tri.V[0] = triangle.V[1];
				tri.V[1] = triangle1.V[1];
				tri.V[2] = triangle1.V[2];

				if (IntrLine3Triangle3_Find(line, tri, TriBary)) {
					ClosestContactTime = min(ClosestContactTime, TriBary[1] + TriBary[2]);
				}
			}

			{
				Triangle3 tri;
				tri.V[0] = triangle.V[1];
				tri.V[1] = triangle.V[2];
				tri.V[2] = triangle1.V[2];

				if (IntrLine3Triangle3_Find(line, tri, TriBary)) {
					ClosestContactTime = min(ClosestContactTime, TriBary[2]);
				}
			}

			{
				Triangle3 tri;
				tri.V[0] = triangle.V[2];
				tri.V[1] = triangle1.V[2];
				tri.V[2] = triangle1.V[0];

				if (IntrLine3Triangle3_Find(line, tri, TriBary)) {
					ClosestContactTime = min(ClosestContactTime, TriBary[1] + TriBary[2]);
				}
			}

			{
				Triangle3 tri;
				tri.V[0] = triangle.V[2];
				tri.V[1] = triangle.V[0];
				tri.V[2] = triangle1.V[0];

				if (IntrLine3Triangle3_Find(line, tri, TriBary)) {
					ClosestContactTime = min(ClosestContactTime, TriBary[2]);
				}
			}
		}

		if (2.0 != ClosestContactTime)
		{
			ContactTime = tmax * ClosestContactTime;
			return true;
		}
		else
		{
			return false;
		}
	}
}

void main() {
	// Shade all the fragments behind the z-buffer
	/*gl_FragColor = vec4(sin(verpos.x*50.0), sin(verpos.y*50.0), 1.0 + 0.0*sin(verpos.z*5.0), 1);
	return;*/

	/*Line3 line; line.Origin = vec3(verpos.x, verpos.y, -1); line.Direction = vec3(0, 0, 1);
	Triangle3 triangle; triangle.V[0] = tri0v0; triangle.V[1] = tri0v1; triangle.V[2] = tri0v2;
	float triBary[3];
	if (IntrLine3Triangle3_Find(line, triangle, triBary))
	{
		gl_FragColor = vec4(triBary[0], triBary[1], triBary[2], 1);
	}
	else discard;
	return;*/

	Line3 line; line.Origin = vec3(verpos.x, verpos.y, -1); line.Direction = vec3(0, 0, 1);
	Triangle3 triangle0; triangle0.V[0] = tri0v0; triangle0.V[1] = tri0v1; triangle0.V[2] = tri0v2;
	Triangle3 triangle1; triangle1.V[0] = tri1v0; triangle1.V[1] = tri1v1; triangle1.V[2] = tri1v2;
	float ContactTime;

	/*gl_FragColor = vec4(0.0, 0.0, 0.0, 1.0);
	if (IntrLine3Triangle3_Find(line, triangle0, 1.0, vec3(0.0), vec3(triangle1.V[0] - triangle0.V[0]), ContactTime))
	{
		//gl_FragColor = vec4(1.0 - ContactTime, 1.0 - ContactTime, 1.0 - ContactTime, 1.0);
		gl_FragColor.g = 1.0;
	}
	else gl_FragColor.r = 1.0;
	return;*/

	bool col = IntrLine3Triangle3_Find(line, triangle0, 1.0, vec3(0.0), vec3(triangle1.V[0] - triangle0.V[0]), ContactTime);

	if (col)
	{
		float t0 = ContactTime;

		if (IntrLine3Triangle3_Find(line, triangle1, 1.0, vec3(0.0), vec3(triangle0.V[0] - triangle1.V[0]), ContactTime))
		{
			float t1 = ContactTime;

			//gl_FragColor = vec4(1.0 - t0 - t1, 1.0 - t0 - t1, 1.0 - t0 - t1, 1.0);
			gl_FragColor = vec4(0.8, 0.3, 0.01, 1.0 - t0 - t1);
		}
		else
			//gl_FragColor = vec4(0.0, 1.0, 0.0, 1.0);
			discard;
	}
	else
		//gl_FragColor = vec4(1.0, 0.0, 0.0, 1.0);
		discard;
}

/*void main() {
	gl_FragColor = vec4(0.8, 0.3, 0.01, 1.0);
}*/
`
